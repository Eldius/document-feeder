package cmd

import (
	"fmt"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/spf13/cobra"
)

// modelsAutoconfigureCmd represents the autoconfigure command
var modelsAutoconfigureCmd = &cobra.Command{
	Use:   "autoconfigure",
	Short: "Fetch model definitions and set configuration patterns for this model",
	Long:  `Fetch model definitions and set configuration patterns for this model.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := ollama.NewOllamaClient()

		res, err := c.ModelDetails(cmd.Context(), modelsAutoconfigureOpts.model)
		if err != nil {
			cmd.PrintErrf("Failed to fetch model details: %v", err)
			return
		}

		contextLength := res.ContextLength()
		fmt.Printf("---\ncontext length: %d\n\n", contextLength)

		chunkSize := contextLength / 2
		chunkOverlap := chunkSize / 10

		fmt.Printf("---\nmodel set to %s\n", modelsAutoconfigureOpts.model)
		fmt.Printf("chunk size: %d\nchunck overlap: %d\n\n", chunkSize, chunkOverlap)

		config.SetOllamaEmbeddingModel(modelsAutoconfigureOpts.model)
		config.SetOllamaEmbeddingChunkSize(chunkSize)
		config.SetOllamaEmbeddingChunkOverlap(chunkOverlap)
		if err := config.PersistConfig(); err != nil {
			cmd.PrintErrf("Failed to persist configuration: %v", err)
			return
		}
	},
}

var (
	modelsAutoconfigureOpts struct {
		model string
	}
)

func init() {
	modelsCmd.AddCommand(modelsAutoconfigureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelsAutoconfigureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modelsAutoconfigureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	modelsAutoconfigureCmd.Flags().StringVarP(&modelsAutoconfigureOpts.model, "model", "m", "", "model to autoconfigure")
}
