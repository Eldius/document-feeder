package adapter

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/ui"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"image/color"
)

const (
	durationMetricNameLabel    = "duration_ms"
	contextSizeMetricNameLabel = "context_size"
	metricNameLabel            = "__name__"
	metricModelLabel           = "model"
)

type Benchmark struct {
	c         ollama.Client
	db        *tsdb.DB
	iterCount int
}

func NewTSDB(path string) (*tsdb.DB, error) {
	_ = os.MkdirAll(path, 0755)
	return tsdb.Open(path, slog.With("pkg", "tsdb"), nil, tsdb.DefaultOptions(), nil)
}

func NewBenchmark(c ollama.Client, db *tsdb.DB, iterCount int) *Benchmark {
	return &Benchmark{c: c, db: db, iterCount: iterCount}
}

func NewBenchmarkFromConfig() (*Benchmark, error) {
	// Create a random dir to work in.  Open() doesn't require a pre-existing dir, but
	// we want to make sure not to make a mess where we shouldn't.
	// Open a TSDB for reading and/or writing.
	db, err := NewTSDB("data/tsdb.db")
	if err != nil {
		return nil, fmt.Errorf("opening TSDB: %w", err)
	}

	return NewBenchmark(ollama.NewOllamaClient(), db, 3), nil
}

func (b *Benchmark) Run(ctx context.Context, models []string) error {
	questionsList := []string{
		"Explique a diferença entre aprendizado supervisionado e não supervisionado.",
		"Liste as principais vantagens do Raspberry Pi 5.",
	}

	f, err := os.OpenFile("data/benchmark.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening benchmark file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"Model", "Question", "Answer", "Time"}); err != nil {
		return fmt.Errorf("writing benchmark file: %w", err)
	}

	if err := CSVHeader(w); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	for range b.iterCount {
		for _, m := range models {
			for _, q := range questionsList {
				if _, err := execute(ctx, b.db, b.c, m, q, w); err != nil {
					return fmt.Errorf("executing benchmark: %w", err)
				}
			}
		}
	}

	return b.Plot(ctx, models)
}

func execute(ctx context.Context, db *tsdb.DB, c ollama.Client, model, question string, w *csv.Writer) (*BenchmarkResult, error) {
	app := db.Appender(ctx)
	labelsGenerationTime := labels.FromStrings(metricNameLabel, durationMetricNameLabel, "model", model, "question", question)
	labelsTokenCount := labels.FromStrings(metricNameLabel, contextSizeMetricNameLabel, "model", model, "question", question)
	refGenerationDuration, err := app.Append(
		0,
		labelsGenerationTime,
		time.Now().Unix(),
		0,
	)
	if err != nil {
		return nil, err
	}
	refGenerationTokenCount, err := app.Append(
		0,
		labelsTokenCount,
		time.Now().Unix(),
		0,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = app.Rollback()
	}()

	//for range iterCount {
	result, err := generate(ctx, c, model, question)
	result.Labels.Duration = labelsGenerationTime
	result.Labels.TokenCount = labelsTokenCount

	if err != nil {
		return nil, fmt.Errorf("generating benchmark result: %w", err)
	}
	_, err = app.Append(refGenerationDuration, result.Labels.Duration, time.Now().Unix(), float64(result.Duration.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("[generation time] appending benchmark result: %w", err)
	}
	_, err = app.Append(refGenerationTokenCount, result.Labels.TokenCount, time.Now().Unix(), float64(result.ContextSize))
	if err != nil {
		return nil, fmt.Errorf("[context size] appending benchmark result: %w", err)
	}

	if err := result.WriteCSVRecord(w); err != nil {
		return nil, fmt.Errorf("writing CSV: %w", err)
	}
	//}
	if err := app.Commit(); err != nil {
		return nil, fmt.Errorf("committing benchmark results: %w", err)
	}

	return nil, nil
}

func generate(ctx context.Context, c ollama.Client, model, question string) (*BenchmarkResult, error) {
	cancel := ui.ProcessingScreen(ctx, fmt.Sprintf("Processing model %s with question %s", model, question))
	defer cancel()

	start := time.Now()
	res, err := c.GenerateCall(ctx, ollama.GenerateRequest{
		Model:     model,
		Prompt:    question,
		Stream:    false,
		KeepAlive: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("generating response: %w", err)
	}

	result := BenchmarkResult{
		Labels: MetricsLabels{
			Duration:   labels.FromStrings(metricNameLabel, durationMetricNameLabel, "model", model, "question", question),
			TokenCount: labels.FromStrings(metricNameLabel, contextSizeMetricNameLabel, "model", model, "question", question),
		},
		Operation:   GenerateOperation,
		Duration:    time.Since(start),
		Model:       model,
		Question:    question,
		Answer:      res.Response,
		ContextSize: len(res.Context),
	}

	fmt.Printf("---\nModel: %s, Question: %s, Answer: %s, Time: %s\n", result.Model, result.Question, result.Answer, result.Duration)

	return &result, nil
}

func (b *Benchmark) Plot(ctx context.Context, models []string) error {
	now := time.Now()
	q, err := b.db.Querier(now.Add(-24*time.Hour).Unix(), now.Unix())
	if err != nil {
		return fmt.Errorf("creating querier: %w", err)
	}
	defer func() { _ = q.Close() }()

	getSeries := func(metricName, model string) plotter.XYs {
		seriesSet := q.Select(
			ctx,
			true,
			nil,
			labels.MustNewMatcher(labels.MatchEqual, metricNameLabel, metricName),
			labels.MustNewMatcher(labels.MatchRegexp, metricModelLabel, model),
		)
		points := plotter.XYs{}
		for seriesSet.Next() {
			it := seriesSet.At().Iterator(nil)
			for it.Next() == chunkenc.ValFloat {
				t, v := it.At()
				points = append(points, plotter.XY{
					X: float64(t) / 1000.0,
					Y: v,
				})
			}
		}
		return points
	}

	duration := make(map[string]plotter.XYs)
	tokens := make(map[string]plotter.XYs)
	for _, m := range models {
		duration[m] = getSeries(durationMetricNameLabel, m)
		tokens[m] = getSeries(contextSizeMetricNameLabel, m)
	}

	p := plot.New()
	p.Title.Text = "Métricas do sistema"
	p.X.Label.Text = "Tempo (s)"
	p.Y.Label.Text = "Valor"

	for m, _ := range duration {
		durationVals := duration[m]
		tokensVals := tokens[m]
		durLine, err := plotter.NewLine(durationVals)
		if err != nil {
			return fmt.Errorf("creating duration line: %w", err)
		}
		durLine.LineStyle.Color = randonColor()
		durLine.LineStyle.Width = vg.Points(1)

		p.Add(durLine)
		p.Legend.Add(fmt.Sprintf("%s - duration (ms)", m), durLine)

		tokensLine, err := plotter.NewLine(tokensVals)
		if err != nil {
			return fmt.Errorf("creating duration line: %w", err)
		}
		tokensLine.LineStyle.Color = randonColor()
		tokensLine.LineStyle.Width = vg.Points(1)
		p.Add(tokensLine)
		p.Legend.Add(fmt.Sprintf("%s - tokens", m), tokensLine)

		p.Legend.Top = true
	}
	if err := p.Save(8*vg.Inch, 4*vg.Inch, "metrics.png"); err != nil {
		return fmt.Errorf("saving plot: %w", err)
	}

	return nil
}

func randonColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.IntN(255)),
		G: uint8(rand.IntN(255)),
		B: uint8(rand.IntN(255)),
		A: uint8(rand.IntN(255)),
	}
}

type Operation int

func (o Operation) String() string {
	switch o {
	case GenerateOperation:
		return "generate"
	case EmbeddingsOperation:
		return "embeddings"
	}
	return ""
}

const (
	GenerateOperation Operation = iota
	EmbeddingsOperation
)

type BenchmarkResult struct {
	Labels      MetricsLabels
	Duration    time.Duration
	Operation   Operation
	Model       string
	Question    string
	Answer      string
	ContextSize int
}

type MetricsLabels struct {
	Duration   labels.Labels
	TokenCount labels.Labels
}

func (r BenchmarkResult) WriteCSVRecord(w *csv.Writer) error {
	return w.Write([]string{
		r.Operation.String(),
		r.Model,
		r.Question,
		r.Answer,
		fmt.Sprintf("%d", r.Duration.Milliseconds()),
		fmt.Sprintf("%d", r.ContextSize),
		r.Labels.Duration.String(),
		r.Labels.TokenCount.String(),
	})
}

func CSVHeader(w *csv.Writer) error {
	return w.Write([]string{
		"operation",
		"model",
		"question",
		"answer",
		"duration_ms",
		"context_size",
		"generation_time_labels",
		"context_size_labels",
	})
}
