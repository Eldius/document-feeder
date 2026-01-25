package cmd

import (
	"github.com/eldius/document-feed-embedder/internal/config"
	"github.com/eldius/initial-config-go/configs"
	"os"

	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "document-feed-embedder",
	Short: "A simple news feed tool",
	Long:  `A simple news feed tool.`,
	PersistentPreRunE: setup.PersistentPreRunE(
		"feed-embedder",
		setup.WithConfigFileToBeUsed(cfgFile),
		setup.WithDefaultCfgFileLocations("~", ".config", "."),
		setup.WithEnvPrefix("FEEDER"),
		setup.WithDefaultCfgFileName("config"),
		setup.WithDefaultValues(configs.DefaultConfigValuesLogFileMap),
		setup.WithProps(
			config.OllamaEndPointProp,
			config.OllamaEmbeddingModelProp,
			config.OllamaEmbeddingBatchSizeProp,
			config.OllamaEmbeddingChunkOverlapProp,
			config.OllamaGenerationModelProp,
			config.OllamaGenerationCacheEnabledProp,
		),
	),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	cfgFile string
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.document-feed-embedder.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
