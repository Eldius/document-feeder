package adapter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/ui"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
)

type Benchmark struct {
	c  ollama.Client
	db *tsdb.DB
}

func NewBenchmark(c ollama.Client, db *tsdb.DB) *Benchmark {
	return &Benchmark{c: c, db: db}
}

func NewBenchmarkFromConfig() (*Benchmark, error) {
	// Create a random dir to work in.  Open() doesn't require a pre-existing dir, but
	// we want to make sure not to make a mess where we shouldn't.
	dbPath := "data/tsdb.db"
	_ = os.MkdirAll(dbPath, 0755)

	// Open a TSDB for reading and/or writing.
	db, err := tsdb.Open(dbPath, nil, nil, tsdb.DefaultOptions(), nil)
	if err != nil {
		return nil, fmt.Errorf("opening TSDB: %w", err)
	}

	return NewBenchmark(ollama.NewOllamaClient(), db), nil
}

func (b *Benchmark) Run(ctx context.Context, models []string) error {
	questionsList := []string{
		"Explique a diferença entre aprendizado supervisionado e não supervisionado.",
		"Liste as principais vantagens do Raspberry Pi 5.",
	}

	for _, m := range models {
		for _, q := range questionsList {
			if _, err := execute(ctx, b.db, b.c, m, q); err != nil {
				return fmt.Errorf("executing benchmark: %w", err)
			}
		}
	}
	return nil
}

func execute(ctx context.Context, db *tsdb.DB, c ollama.Client, model, question string) (*BenchmarkResult, error) {
	app := db.Appender(ctx)
	labelsGenerationTime := labels.FromStrings("__name__", "generation_time", "model", model, "question", question)
	refGenerationDuration, err := app.Append(
		0,
		labelsGenerationTime,
		time.Now().Unix(),
		0,
	)
	if err != nil {
		return nil, err
	}
	labelsTokenCount := labels.FromStrings("__name__", "context_size", "model", model, "question", question)
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

	//for range 10 {
	result, err := generate(ctx, c, model, question)
	if err != nil {
		return nil, fmt.Errorf("generating benchmark result: %w", err)
	}
	_, err = app.Append(refGenerationDuration, labelsGenerationTime, time.Now().Unix(), float64(result.Duration.Milliseconds()))
	if err != nil {
		return nil, fmt.Errorf("[generation time] appending benchmark result: %w", err)
	}
	_, err = app.Append(refGenerationTokenCount, labelsTokenCount, time.Now().Unix(), float64(result.ContextSize))
	if err != nil {
		return nil, fmt.Errorf("[context size] appending benchmark result: %w", err)
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
	Operation   Operation
	Duration    time.Duration
	Model       string
	Question    string
	Answer      string
	ContextSize int
}
