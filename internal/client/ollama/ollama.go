package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/initial-config-go/httpclient"
)

type OllamaClient interface {
	EmbeddingFuncSingleShot(ctx context.Context, text string) ([]float32, error)
	EmbeddingFuncKeepAlive(ctx context.Context, text string) ([]float32, error)
	EmbeddingFunc(ctx context.Context, text string) ([]float32, error)
	EmbeddingCall(ctx context.Context, reqPayload OllamaEmbeddingRequest) (*OllamaEmbeddingResponse, error)
	ChatFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (*OllamaChatResponse, error)
	GenerateFunc(ctx context.Context, prompt string, opts ...GenerationOption) (*OllamaGenerateResponse, error)
	ListModels(ctx context.Context) (*OllamaModelsResponse, error)
	RunningModels(ctx context.Context) (*OllamaModelsResponse, error)
	ModelDetailsCall(ctx context.Context, payload OllamaModelDetailsRequest) (*OllamaModelDetailsResponse, error)
	ModelDetails(ctx context.Context, model string) (*OllamaModelDetailsResponse, error)
}

type client struct {
	c                  *http.Client
	endpoint           string
	embeddingModel     string
	embeddingBatchSize int
	generationModel    string
}

func NewOllamaClient() OllamaClient {
	return &client{
		c:                  httpclient.NewHTTPClient(),
		endpoint:           config.GetOllamaEndpoint(),
		embeddingModel:     config.GetOllamaEmbeddingModel(),
		embeddingBatchSize: config.GetOllamaEmbeddingChunkSize(),
		generationModel:    config.GetOllamaGenerationModel(),
	}
}

func (c *client) EmbeddingFuncSingleShot(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, OllamaEmbeddingRequest{
		Model:     c.embeddingModel,
		Input:     []string{text},
		KeepAlive: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *client) EmbeddingFuncKeepAlive(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, OllamaEmbeddingRequest{
		Model:     c.embeddingModel,
		Input:     []string{text},
		KeepAlive: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *client) EmbeddingFunc(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, OllamaEmbeddingRequest{
		Model: c.embeddingModel,
		Input: []string{text},
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *client) EmbeddingCall(ctx context.Context, reqPayload OllamaEmbeddingRequest) (*OllamaEmbeddingResponse, error) {
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
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	var embeddings OllamaEmbeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &embeddings, nil
}

func WithNumKeep(numKeep int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumKeep = numKeep
	}
}

func WithSeed(seed int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.Seed = seed
	}
}

func WithNumPredict(numPredict int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumPredict = numPredict
	}
}

func WithTopK(topK int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.TopK = topK
	}
}

func WithTopP(topP float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.TopP = topP
	}
}

func WithMinP(minP float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.MinP = minP
	}
}

func WithTypicalP(typicalP float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.TypicalP = typicalP
	}
}

func WithRepeatLastN(repeatLastN int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.RepeatLastN = repeatLastN
	}
}

func WithTemperature(temperature float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.Temperature = temperature
	}
}

func WithRepeatPenalty(repeatPenalty float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.RepeatPenalty = repeatPenalty
	}
}

func WithPresencePenalty(presencePenalty float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.PresencePenalty = presencePenalty
	}
}

func WithFrequencyPenalty(frequencyPenalty float64) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.FrequencyPenalty = frequencyPenalty
	}
}

func WithPenalizeNewline(penalizeNewline bool) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.PenalizeNewline = penalizeNewline
	}
}

func WithStop(stop []string) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.Stop = stop
	}
}

func WithNuma(numa bool) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.Numa = numa
	}
}

func WithNumCtx(numCtx int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumCtx = numCtx
	}
}

func WithNumBatch(numBatch int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumBatch = numBatch
	}
}

func WithNumGpu(numGpu int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumGpu = numGpu
	}
}

func WithMainGpu(mainGpu int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.MainGpu = mainGpu
	}
}

func WithUseMmap(useMmap bool) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.UseMmap = useMmap
	}
}

func WithNumThread(numThread int) GenerationOption {
	return func(opts *ollamaOptionsRequest) {
		opts.NumThread = numThread
	}
}

func (c *client) ChatFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (*OllamaChatResponse, error) {
	options := defaultOllamaGenerationOptions()
	for _, opt := range opts {
		opt(&options)
	}

	reqPayload := ollamaChatRequest{
		Model: c.generationModel,
		Messages: []OllamaChatMessage{{
			Role:    "user",
			Content: prompt,
		}},
		Stream:  false,
		Think:   think,
		Options: options,
	}

	b, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/chat", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response OllamaChatResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &response, nil
}

func (c *client) GenerateFunc(ctx context.Context, prompt string, opts ...GenerationOption) (*OllamaGenerateResponse, error) {
	options := defaultOllamaGenerationOptions()
	for _, opt := range opts {
		opt(&options)
	}

	reqPayload := ollamaGenerateRequest{
		Model:   c.generationModel,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	b, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/generate", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response OllamaGenerateResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &response, nil
}

func (c *client) ListModels(ctx context.Context) (*OllamaModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response OllamaModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *client) RunningModels(ctx context.Context) (*OllamaModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/api/ps", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response OllamaModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *client) ModelDetailsCall(ctx context.Context, payload OllamaModelDetailsRequest) (*OllamaModelDetailsResponse, error) {
	b, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/show", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response OllamaModelDetailsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *client) ModelDetails(ctx context.Context, model string) (*OllamaModelDetailsResponse, error) {
	return c.ModelDetailsCall(ctx, OllamaModelDetailsRequest{Model: model})
}
