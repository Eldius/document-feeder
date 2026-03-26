package cmd

import (
	"os"
	"time"

	"github.com/eldius/initial-config-go/telemetry"
	"github.com/spf13/viper"

	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/initial-config-go/configs"

	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "document-feed-embedder",
	Short: "A simple news feed tool",
	Long:  `A simple news feed tool.`,
	PersistentPreRunE: setup.PersistentPreRunE(
		config.CliAppName,
		setup.WithConfigFileToBeUsed(rootOpts.cfgFile),
		setup.WithDefaultCfgFileLocations("~", ".config", "."),
		setup.WithEnvPrefix("FEEDER"),
		setup.WithDefaultCfgFileName("config"),
		setup.WithDefaultValues(configs.DefaultConfigValuesLogFileMap),
		setup.WithOpenTelemetryOptions(telemetry.WithService(config.CliAppName, config.Version, "")),
		setup.WithProps(
			config.OllamaEndPointProp,
			config.OllamaEmbeddingModelProp,
			config.OllamaEmbeddingChunkSizeProp,
			config.OllamaEmbeddingChunkOverlapProp,
			config.OllamaGenerationModelProp,
			config.OllamaGenerationCacheEnabledProp,
			config.OllamaGenerationNoCacheProp,
			config.XmppNotifierURLProp,
			config.XmppNotifierUserProp,
			config.XmppNotifierPassProp,
			config.XmppNotifierRecipientProp,
			config.XmppNotifierEnabledProp,
		),
	),
	PersistentPostRunE: setup.PersistentPostRunE(1 * time.Second),
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
	rootOpts struct {
		enableNotification bool
		cfgFile            string
	}
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&rootOpts.cfgFile, "config", "", "config file (default is $HOME/.document-feed-embedder.yaml)")
	rootCmd.PersistentFlags().BoolVar(&rootOpts.enableNotification, "enable-notification", false, "Enable notification after execution")
	_ = viper.BindPFlag(config.XmppNotifierEnabledProp.Key, rootCmd.PersistentFlags().Lookup("enable-notification"))

}
