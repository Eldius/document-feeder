package config

import (
	"os"
	"time"

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
	OllamaEmbeddingChunkSizeProp = setup.Prop{
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
	OllamaGenerationNoCacheProp = setup.Prop{
		Key:   "ollama.generation.no-cache",
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

func GetOllamaEmbeddingChunkSize() int {
	return viper.GetInt(OllamaEmbeddingChunkSizeProp.Key)
}

func GetOllamaEmbeddingChunkOverlap() int {
	return viper.GetInt(OllamaEmbeddingChunkOverlapProp.Key)
}

func GetOllamaGenerationModel() string {
	return viper.GetString(OllamaGenerationModelProp.Key)
}

func GetOllamaGenerationCacheEnabled() bool {
	return viper.GetBool(OllamaGenerationCacheEnabledProp.Key) && !viper.GetBool(OllamaGenerationNoCacheProp.Key)
}

func GetOllamaGenerationCacheSimilarityThreshold() float32 {
	return float32(viper.GetFloat64(OllamaGenerationCacheSimilarityThresholdProp.Key))
}

func SetOllamaEmbeddingModel(model string) {
	viper.Set(OllamaEmbeddingModelProp.Key, model)
}

func SetOllamaEmbeddingChunkSize(chunkSize int) {
	viper.Set(OllamaEmbeddingChunkSizeProp.Key, chunkSize)
}

func SetOllamaEmbeddingChunkOverlap(chunkOverlap int) {
	viper.Set(OllamaEmbeddingChunkOverlapProp.Key, chunkOverlap)
}

func PersistConfig() error {
	cfgFile := viper.ConfigFileUsed()
	curCffFileContent, err := os.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfgFile+"."+time.Now().Format("2006-01-02_15-04-05")+".yaml", curCffFileContent, 0644); err != nil {
		return err
	}

	return viper.WriteConfig()
}

func GetFetchConfigStruct(key string, val any) error {
	return viper.UnmarshalKey(key, &val)
}
