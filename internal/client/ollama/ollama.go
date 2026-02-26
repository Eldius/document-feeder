package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/otel/metric"
	"net/http"

	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/initial-config-go/http/client"
)

type Client interface {
	EmbeddingFuncSingleShot(ctx context.Context, text string) ([]float32, error)
	EmbeddingFuncKeepAlive(ctx context.Context, text string) ([]float32, error)
	EmbeddingFunc(ctx context.Context, text string) ([]float32, error)
	EmbeddingCall(ctx context.Context, reqPayload EmbeddingRequest) (*EmbeddingResponse, error)
	ChatFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (*ChatResponse, error)
	GenerateFunc(ctx context.Context, prompt string, opts ...GenerationOption) (*GenerateResponse, error)
	ListModels(ctx context.Context) (*ModelsResponse, error)
	RunningModels(ctx context.Context) (*ModelsResponse, error)
	ModelDetailsCall(ctx context.Context, payload ModelDetailsRequest) (*ModelDetailsResponse, error)
	ModelDetails(ctx context.Context, model string) (*ModelDetailsResponse, error)
	GenerateCall(ctx context.Context, reqPayload GenerateRequest) (*GenerateResponse, error)
	GenerateCallStream(ctx context.Context, ch chan string, reqPayload GenerateRequest) error
}

type ollamaClient struct {
	c                     *http.Client
	endpoint              string
	embeddingModelName    string
	embeddingChunkSize    int
	generationModelName   string
	meter                 metric.Meter
	promptTokenCounter    metric.Int64Counter
	generatedTokenCounter metric.Int64Counter
}

func NewOllamaClientFromConfigs() (Client, error) {
	meter := telemetry.GetMeter("ollama")
	tokenCountMeter, err := meter.Int64Counter("prompt_token_count", metric.WithDescription(""))
	if err != nil {
		err = fmt.Errorf("creating token count meter: %w", err)
		return nil, err
	}
	tokenGenerateCounter, err := meter.Int64Counter("generated_token_count", metric.WithDescription(""))
	if err != nil {
		err = fmt.Errorf("creating token generate meter: %w", err)
		return nil, err
	}
	return NewOllamaClient(
		client.NewHTTPClient(),
		config.GetOllamaEndpoint(),
		config.GetOllamaEmbeddingModel(),
		config.GetOllamaGenerationModel(),
		meter,
		tokenCountMeter,
		tokenGenerateCounter,
		config.GetOllamaEmbeddingChunkSize(),
	), nil
}

func NewOllamaClient(c *http.Client, endpoint, embeddingModel, generationModel string, meter metric.Meter, tokenCounter, tokenGenerateCounter metric.Int64Counter, embeddingChunkSize int) Client {
	return &ollamaClient{
		c:                     c,
		endpoint:              endpoint,
		embeddingModelName:    embeddingModel,
		generationModelName:   generationModel,
		embeddingChunkSize:    embeddingChunkSize,
		meter:                 meter,
		promptTokenCounter:    tokenCounter,
		generatedTokenCounter: tokenGenerateCounter,
	}
}

func (c *ollamaClient) EmbeddingFuncSingleShot(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, EmbeddingRequest{
		Model:     c.embeddingModelName,
		Input:     []string{text},
		KeepAlive: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *ollamaClient) EmbeddingFuncKeepAlive(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, EmbeddingRequest{
		Model:     c.embeddingModelName,
		Input:     []string{text},
		KeepAlive: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *ollamaClient) EmbeddingFunc(ctx context.Context, text string) ([]float32, error) {
	res, err := c.EmbeddingCall(ctx, EmbeddingRequest{
		Model: c.embeddingModelName,
		Input: []string{text},
	})
	if err != nil {
		return nil, fmt.Errorf("embedding call failed: %w", err)
	}
	return res.Embeddings[0], err
}

func (c *ollamaClient) EmbeddingCall(ctx context.Context, reqPayload EmbeddingRequest) (*EmbeddingResponse, error) {
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
	var embeddings EmbeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &embeddings, nil
}

func (c *ollamaClient) ChatFunc(ctx context.Context, prompt string, think bool, opts ...GenerationOption) (*ChatResponse, error) {
	options := defaultOllamaGenerationOptions()
	for _, opt := range opts {
		opt(&options)
	}

	reqPayload := ChatRequest{
		Model: c.generationModelName,
		Messages: []ChatMessage{{
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

	var response ChatResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &response, nil
}

func (c *ollamaClient) GenerateCall(ctx context.Context, reqPayload GenerateRequest) (*GenerateResponse, error) {
	reqPayload.Stream = false
	if reqPayload.Model == "" {
		reqPayload.Model = c.generationModelName
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

	var response GenerateResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.promptTokenCounter.Add(ctx, response.PromptEvalCount)

	return &response, nil
}

func (c *ollamaClient) GenerateCallStream(ctx context.Context, ch chan string, reqPayload GenerateRequest) error {
	reqPayload.Stream = true
	if reqPayload.Model == "" {
		reqPayload.Model = c.generationModelName
	}

	b, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("marshalling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/generate", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		chunkBytes := scanner.Bytes()
		logs.NewLogger(ctx, logs.KeyValueData{
			"chunk": string(chunkBytes),
		}).Debug("chunk")

		var chunk GenerateResponse
		if err := json.Unmarshal(chunkBytes, &chunk); err != nil {
			return fmt.Errorf("unmarshalling chunk: %w", err)
		}

		ch <- chunk.Response
		if chunk.Done {
			break
		}
	}

	return nil
}

func (c *ollamaClient) GenerateFunc(ctx context.Context, prompt string, opts ...GenerationOption) (*GenerateResponse, error) {
	options := defaultOllamaGenerationOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return c.GenerateCall(ctx, GenerateRequest{
		Model:   c.generationModelName,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	})
}

func (c *ollamaClient) ListModels(ctx context.Context) (*ModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response ModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *ollamaClient) RunningModels(ctx context.Context) (*ModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/api/ps", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response ModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *ollamaClient) ModelDetailsCall(ctx context.Context, payload ModelDetailsRequest) (*ModelDetailsResponse, error) {
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

	var response ModelDetailsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}

func (c *ollamaClient) ModelDetails(ctx context.Context, model string) (*ModelDetailsResponse, error) {
	return c.ModelDetailsCall(ctx, ModelDetailsRequest{Model: model})
}
