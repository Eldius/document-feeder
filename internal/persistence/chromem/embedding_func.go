package chromem

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/initial-config-go/logs"
	"github.com/philippgille/chromem-go"
)

const (
	isNormalizedPrecisionTolerance = 1e-6
)

// NewEmbeddingFuncOllama returns a function that creates embeddings for a text
// using Ollama's embedding API. You can pass any model that Ollama supports and
// that supports embeddings. A good one as of 2024-03-02 is "nomic-embed-text".
// See https://ollama.com/library/nomic-embed-text
// baseURLOllama is the base URL of the Ollama API. If it's empty,
// "http://localhost:11434/api" is used.
func NewEmbeddingFuncOllama(model string, oc ollama.Client) chromem.EmbeddingFunc {

	var checkedNormalized bool
	checkNormalized := sync.Once{}

	return func(ctx context.Context, text string) ([]float32, error) {
		log := logs.NewLogger(ctx, logs.KeyValueData{
			"model": model,
			"text":  text,
		})
		// Prepare the request body.
		reqBody := ollama.EmbeddingRequest{
			Model:     model,
			Input:     []string{text},
			KeepAlive: 60,
		}

		log.Debug("Calling embedding API")
		embeddingResponse, err := oc.EmbeddingCall(ctx, reqBody)
		if err != nil {
			return nil, fmt.Errorf("embedding call: %w", err)
		}
		log.WithExtraData("embeddings", embeddingResponse).Debug("Embedding API call successful")

		// Check if the response contains embeddings.
		if len(embeddingResponse.Embeddings) == 0 {
			return nil, errors.New("no embeddings found in the response")
		}

		v := embeddingResponse.Embeddings[0]
		checkNormalized.Do(func() {
			if isNormalized(v) {
				checkedNormalized = true
			} else {
				checkedNormalized = false
			}
		})
		if !checkedNormalized {
			v = normalizeVector(v)
		}

		return v, nil
	}
}

// isNormalized checks if the vector is normalized.
func isNormalized(v []float32) bool {
	var sqSum float64
	for _, val := range v {
		sqSum += float64(val) * float64(val)
	}
	magnitude := math.Sqrt(sqSum)
	return math.Abs(magnitude-1) < isNormalizedPrecisionTolerance
}

func normalizeVector(v []float32) []float32 {
	var norm float32
	for _, val := range v {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	res := make([]float32, len(v))
	for i, val := range v {
		res[i] = val / norm
	}

	return res
}
