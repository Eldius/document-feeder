package adapter

import "github.com/eldius/document-feed-embedder/internal/client"

type OllamaAdapter struct {
	ollama *client.OllamaClient
}

func NewOllamaAdapter(ollama *client.OllamaClient) *OllamaAdapter {
	return &OllamaAdapter{ollama: ollama}
}
