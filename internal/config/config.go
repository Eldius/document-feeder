package config

import "github.com/spf13/viper"

func GetOllamaEndpoint() string {
	return viper.GetString("ollama.endpoint")
}

func GetOllamaEmbeddingModel() string {
	return viper.GetString("ollama.embedding.model")
}

func GetOllamaEmbeddingBatchSize() int {
	return viper.GetInt("ollama.embedding.chunk_size")
}

func GetOllamaEmbeddingChunkOverlap() int {
	return viper.GetInt("ollama.embedding.chunk_overlap")
}
