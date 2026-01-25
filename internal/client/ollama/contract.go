package ollama

import (
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

type ollamaChatRequest struct {
	Model    string               `json:"model"`
	Messages []OllamaChatMessage  `json:"messages"`
	Stream   bool                 `json:"stream"`
	Think    bool                 `json:"think"`
	Options  ollamaOptionsRequest `|json:"options,omitempty"`
}

type OllamaChatMessage struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

type ollamaOptionsRequest struct {
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

func defaultOllamaGenerationOptions() ollamaOptionsRequest {
	return ollamaOptionsRequest{
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

type OllamaChatResponse struct {
	Model              string            `json:"model"`
	CreatedAt          time.Time         `json:"created_at"`
	Message            OllamaChatMessage `json:"message"`
	DoneReason         string            `json:"done_reason"`
	Done               bool              `json:"done"`
	TotalDuration      int64             `json:"total_duration"`
	LoadDuration       int64             `json:"load_duration"`
	PromptEvalCount    int               `json:"prompt_eval_count"`
	PromptEvalDuration int64             `json:"prompt_eval_duration"`
	EvalCount          int               `json:"eval_count"`
	EvalDuration       int64             `json:"eval_duration"`
}

type OllamaGenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

type ollamaGenerateRequest struct {
	Model   string               `json:"model"`
	Prompt  string               `json:"prompt"`
	Stream  bool                 `json:"stream"`
	Options ollamaOptionsRequest `json:"options"`
}

type GenerationOption func(*ollamaOptionsRequest)

type OllamaModelsResponse struct {
	Models []OllamaModel `json:"models"`
}

type OllamaModel struct {
	Name       string             `json:"name"`
	ModifiedAt time.Time          `json:"modified_at"`
	Size       int64              `json:"size"`
	Digest     string             `json:"digest"`
	Details    OllamaModelSummary `json:"details"`
}

type OllamaModelSummary struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}
