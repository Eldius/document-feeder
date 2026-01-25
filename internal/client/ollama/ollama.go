package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (c *OllamaClient) ChatFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (*OllamaChatResponse, error) {
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

func (c *OllamaClient) GenerateFunc(ctx context.Context, prompt string, opts ...GenerationOption) (*OllamaGenerateResponse, error) {
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

func (c *OllamaClient) ListModels(ctx context.Context) (*OllamaModelsResponse, error) {
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

func (c *OllamaClient) RunningModels(ctx context.Context) (*OllamaModelsResponse, error) {
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
