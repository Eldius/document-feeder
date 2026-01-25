package adapter

import (
	"github.com/eldius/document-feed-embedder/internal/client/ollama"
)

type OllamaAdapter struct {
	ollama *ollama.OllamaClient
}

func NewOllamaAdapter(ollama *ollama.OllamaClient) *OllamaAdapter {
	return &OllamaAdapter{ollama: ollama}
}
