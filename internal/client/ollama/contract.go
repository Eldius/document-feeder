package ollama

import (
	"math/rand/v2"

	"time"
)

type EmbeddingRequest struct {
	Model     string   `json:"model"`      //Model name of a model to be used to embed texts.
	Input     []string `json:"input"`      //Input a list of texts to embed.
	KeepAlive int      `json:"keep_alive"` //KeepAlive number of seconds to keep the model loaded in memory.
}

type EmbeddingResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int         `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

type ChatRequest struct {
	Model     string         `json:"model"`
	Messages  []ChatMessage  `json:"messages"`
	Stream    bool           `json:"stream"`
	Think     bool           `json:"think"`
	KeepAlive int            `json:"keep_alive"`
	Options   OptionsRequest `|json:"options,omitempty"`
}

type ChatMessage struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

type OptionsRequest struct {
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

func defaultOllamaGenerationOptions() OptionsRequest {
	randomSeed := rand.IntN(100)
	return OptionsRequest{
		NumKeep:     10,
		Seed:        randomSeed,
		NumPredict:  1,
		TopK:        40,
		TopP:        1.0,
		MinP:        0.0,
		TypicalP:    1.0,
		RepeatLastN: 0,
		Temperature: 0.7,
	}
}

type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	DoneReason         string      `json:"done_reason"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration"`
	LoadDuration       int64       `json:"load_duration"`
	PromptEvalCount    int         `json:"prompt_eval_count"`
	PromptEvalDuration int64       `json:"prompt_eval_duration"`
	EvalCount          int         `json:"eval_count"`
	EvalDuration       int64       `json:"eval_duration"`
}

type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int64   `json:"context"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int64     `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int64     `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

type GenerateRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	Stream    bool           `json:"stream"`
	KeepAlive int            `json:"keep_alive"`
	Options   OptionsRequest `json:"options"`
}

type ModelsResponse struct {
	Models []Model `json:"models"`
}

type Model struct {
	Name       string       `json:"name"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelSummary `json:"details"`
}

type ModelSummary struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type ModelDetailsRequest struct {
	Model   string `json:"model"`
	Verbose bool   `json:"verbose"`
}

type ModelDetailsResponse struct {
	License      string       `json:"license"`
	ModelFile    string       `json:"modelfile"`
	Parameters   string       `json:"parameters"`
	Template     string       `json:"template"`
	Details      ModelDetails `json:"details"`
	ModelInfo    ModelInfo    `json:"model_info"`
	Tensors      []Tensors    `json:"tensors"`
	Capabilities []string     `json:"capabilities"`
	ModifiedAt   string       `json:"modified_at"`
}
type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}
type ModelInfo map[string]any

type Tensors struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Shape []int  `json:"shape"`
}

func (r ModelDetailsResponse) ContextLength() int {
	return int(r.ModelInfo[r.Details.Family+".context_length"].(float64))
}

type GenerationOption func(*OptionsRequest)

func WithNumKeep(numKeep int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumKeep = numKeep
	}
}

func WithSeed(seed int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.Seed = seed
	}
}

func WithNumPredict(numPredict int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumPredict = numPredict
	}
}

func WithTopK(topK int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.TopK = topK
	}
}

func WithTopP(topP float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.TopP = topP
	}
}

func WithMinP(minP float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.MinP = minP
	}
}

func WithTypicalP(typicalP float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.TypicalP = typicalP
	}
}

func WithRepeatLastN(repeatLastN int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.RepeatLastN = repeatLastN
	}
}

func WithTemperature(temperature float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.Temperature = temperature
	}
}

func WithRepeatPenalty(repeatPenalty float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.RepeatPenalty = repeatPenalty
	}
}

func WithPresencePenalty(presencePenalty float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.PresencePenalty = presencePenalty
	}
}

func WithFrequencyPenalty(frequencyPenalty float64) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.FrequencyPenalty = frequencyPenalty
	}
}

func WithPenalizeNewline(penalizeNewline bool) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.PenalizeNewline = penalizeNewline
	}
}

func WithStop(stop []string) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.Stop = stop
	}
}

func WithNuma(numa bool) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.Numa = numa
	}
}

func WithNumCtx(numCtx int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumCtx = numCtx
	}
}

func WithNumBatch(numBatch int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumBatch = numBatch
	}
}

func WithNumGpu(numGpu int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumGpu = numGpu
	}
}

func WithMainGpu(mainGpu int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.MainGpu = mainGpu
	}
}

func WithUseMmap(useMmap bool) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.UseMmap = useMmap
	}
}

func WithNumThread(numThread int) GenerationOption {
	return func(opts *OptionsRequest) {
		opts.NumThread = numThread
	}
}
