package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/config"
	"github.com/eldius/initial-config-go/httpclient"
	"math/rand/v2"
	"net/http"
	"time"
)

type OllamaClient struct {
	c                  *http.Client
	endpoint           string
	embeddingModel     string
	embeddingBatchSize int
	generationModel    string
}

func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		c:                  httpclient.NewHTTPClient(),
		endpoint:           config.GetOllamaEndpoint(),
		embeddingModel:     config.GetOllamaEmbeddingModel(),
		embeddingBatchSize: config.GetOllamaEmbeddingBatchSize(),
		generationModel:    config.GetOllamaGenerationModel(),
	}
}

type ollamaEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbeddingResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int         `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

type ollamaGenerationRequest struct {
	Model    string                    `json:"model"`
	Messages []ollamaGenerationMessage `json:"messages"`
	Stream   bool                      `json:"stream"`
	Think    bool                      `json:"think"`
	Options  ollamaGenerationOptions   `|json:"options,omitempty"`
}

type ollamaGenerationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaGenerationOptions struct {
	NumKeep          int      `json:"num_keep,omitempty"`
	Seed             int      `json:"seed,omitempty"`
	NumPredict       int      `json:"num_predict,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	MinP             float64  `json:"min_p,omitempty"`
	TypicalP         float64  `json:"typical_p,omitempty"`
	RepeatLastN      int      `json:"repeat_last_n,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	RepeatPenalty    float64  `json:"repeat_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PenalizeNewline  bool     `json:"penalize_newline,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Numa             bool     `json:"numa,omitempty"`
	NumCtx           int      `json:"num_ctx,omitempty"`
	NumBatch         int      `json:"num_batch,omitempty"`
	NumGpu           int      `json:"num_gpu,omitempty"`
	MainGpu          int      `json:"main_gpu,omitempty"`
	UseMmap          bool     `json:"use_mmap,omitempty"`
	NumThread        int      `json:"num_thread,omitempty"`
}

func defaultOllamaGenerationOptions() ollamaGenerationOptions {
	return ollamaGenerationOptions{
		NumKeep:     10,
		Seed:        rand.Int(),
		NumPredict:  1,
		TopK:        40,
		TopP:        1.0,
		MinP:        0.0,
		TypicalP:    1.0,
		RepeatLastN: 0,
		Temperature: 0.7,
	}
}

type ollamaGenerationResponse struct {
	Model              string                  `json:"model"`
	CreatedAt          time.Time               `json:"created_at"`
	Message            ollamaGenerationMessage `json:"message"`
	DoneReason         string                  `json:"done_reason"`
	Done               bool                    `json:"done"`
	TotalDuration      int64                   `json:"total_duration"`
	LoadDuration       int64                   `json:"load_duration"`
	PromptEvalCount    int                     `json:"prompt_eval_count"`
	PromptEvalDuration int64                   `json:"prompt_eval_duration"`
	EvalCount          int                     `json:"eval_count"`
	EvalDuration       int64                   `json:"eval_duration"`
}

func (c *OllamaClient) EmbeddingFunc(ctx context.Context, text string) ([]float32, error) {
	reqPayload := ollamaEmbeddingRequest{
		Model: c.embeddingModel,
		Input: []string{text},
	}
	b, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/embed", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	var embeddings ollamaEmbeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return embeddings.Embeddings[0], nil
}

type GenerationOption func(*ollamaGenerationOptions)

func WithNumKeep(numKeep int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumKeep = numKeep
	}
}

func WithSeed(seed int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.Seed = seed
	}
}

func WithNumPredict(numPredict int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumPredict = numPredict
	}
}

func WithTopK(topK int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.TopK = topK
	}
}

func WithTopP(topP float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.TopP = topP
	}
}

func WithMinP(minP float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.MinP = minP
	}
}

func WithTypicalP(typicalP float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.TypicalP = typicalP
	}
}

func WithRepeatLastN(repeatLastN int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.RepeatLastN = repeatLastN
	}
}

func WithTemperature(temperature float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.Temperature = temperature
	}
}

func WithRepeatPenalty(repeatPenalty float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.RepeatPenalty = repeatPenalty
	}
}

func WithPresencePenalty(presencePenalty float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.PresencePenalty = presencePenalty
	}
}

func WithFrequencyPenalty(frequencyPenalty float64) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.FrequencyPenalty = frequencyPenalty
	}
}

func WithPenalizeNewline(penalizeNewline bool) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.PenalizeNewline = penalizeNewline
	}
}

func WithStop(stop []string) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.Stop = stop
	}
}

func WithNuma(numa bool) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.Numa = numa
	}
}

func WithNumCtx(numCtx int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumCtx = numCtx
	}
}

func WithNumBatch(numBatch int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumBatch = numBatch
	}
}

func WithNumGpu(numGpu int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumGpu = numGpu
	}
}

func WithMainGpu(mainGpu int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.MainGpu = mainGpu
	}
}

func WithUseMmap(useMmap bool) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.UseMmap = useMmap
	}
}

func WithNumThread(numThread int) GenerationOption {
	return func(opts *ollamaGenerationOptions) {
		opts.NumThread = numThread
	}
}

func (c *OllamaClient) GenerationFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (string, error) {
	options := defaultOllamaGenerationOptions()
	for _, opt := range opts {
		opt(&options)
	}

	reqPayload := ollamaGenerationRequest{
		Model: c.generationModel,
		Messages: []ollamaGenerationMessage{{
			Role:    "user",
			Content: prompt,
		}},
		Stream:  false,
		Think:   false,
		Options: options,
	}

	b, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("marshalling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/chat", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	var response ollamaGenerationResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return response.Message.Content, nil
}
