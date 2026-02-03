package cmd

import (
	"os"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "document-feed-embedder",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRunE: setup.PersistentPreRunE(
		"document-feeder-nemchmarker",
		setup.WithConfigFileToBeUsed(cfgFile),
		setup.WithDefaultCfgFileLocations("~", ".config", "."),
		setup.WithEnvPrefix("BENCHMARKER"),
		setup.WithDefaultCfgFileName("config"),
		setup.WithDefaultValues(configs.DefaultConfigValuesLogFileMap),
		setup.WithProps(
			config.OllamaEndPointProp,
			config.OllamaEmbeddingModelProp,
			config.OllamaEmbeddingChunkSizeProp,
			config.OllamaEmbeddingChunkOverlapProp,
			config.OllamaGenerationModelProp,
			config.OllamaGenerationCacheEnabledProp,
			config.OllamaGenerationNoCacheProp,
		),
	),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := adapter.NewBenchmarkFromConfig()
		if err != nil {
			return err
		}

		if err := c.Generate(cmd.Context(), rootOpts.models); err != nil {
			return err
		}

		return nil
	},
}

var (
	cfgFile  string
	rootOpts struct {
		models []string
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringSliceVar(&rootOpts.models, "model", []string{"deepseek-r1:1.5b", "llama3:8b-instruct-q4_K_M", "tinyllama:latest"}, "ollama model to use")
}
