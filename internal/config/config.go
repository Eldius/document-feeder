package config

import (
	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/viper"
)

var (
	OllamaEndPointProp = setup.Prop{
		Key:   "ollama.endpoint",
		Value: "http://localhost:11434",
	}
	OllamaEmbeddingModelProp = setup.Prop{
		Key:   "ollama.embedding.model",
		Value: "nomic-embed-text",
	}
	OllamaEmbeddingBatchSizeProp = setup.Prop{
		Key:   "ollama.embedding.chunk_size",
		Value: 2048,
	}
	OllamaEmbeddingChunkOverlapProp = setup.Prop{
		Key:   "ollama.embedding.chunk_overlap",
		Value: 200,
	}
	OllamaGenerationModelProp = setup.Prop{
		Key:   "ollama.generation.model",
		Value: "llama3:8b-instruct-q4_K_M",
	}
	OllamaGenerationCacheEnabledProp = setup.Prop{
		Key:   "ollama.generation.cache_enabled",
		Value: false,
	}
	OllamaGenerationCacheSimilarityThresholdProp = setup.Prop{
		Key:   "ollama.generation.cache_similarity_threshold",
		Value: 0.8,
	}
)

func GetOllamaEndpoint() string {
	return viper.GetString(OllamaEndPointProp.Key)
}

func GetOllamaEmbeddingModel() string {
	return viper.GetString(OllamaEmbeddingModelProp.Key)
}

func GetOllamaEmbeddingBatchSize() int {
	return viper.GetInt(OllamaEmbeddingBatchSizeProp.Key)
}

func GetOllamaEmbeddingChunkOverlap() int {
	return viper.GetInt(OllamaEmbeddingChunkOverlapProp.Key)
}

func GetOllamaGenerationModel() string {
	return viper.GetString(OllamaGenerationModelProp.Key)
}

func GetOllamaGenerationCacheEnabled() bool {
	return viper.GetBool(OllamaGenerationCacheEnabledProp.Key)
}

func GetOllamaGenerationCacheSimilarityThreshold() float32 {
	return float32(viper.GetFloat64(OllamaGenerationCacheSimilarityThresholdProp.Key))
}
