package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/document-feeder/internal/server/api"
	"github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "document-feeder",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: setup.PersistentPreRunE(
		"document-feeder",
		setup.WithConfigFileToBeUsed(rootOpts.cfgFile),
		setup.WithDefaultCfgFileLocations("~", ".config", "."),
		setup.WithEnvPrefix("FEEDER"),
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
			config.XmppNotifierURLProp,
			config.XmppNotifierUserProp,
			config.XmppNotifierPassProp,
			config.XmppNotifierRecipientProp,
			config.XmppNotifierEnabledProp,
			config.ApiPortProp,
		),
	),
	PersistentPostRunE: setup.PersistentPostRunE(1 * time.Second),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := api.StartServer(cmd.Context(), config.GetApiPort()); err != nil {
			fmt.Printf("failed to start server: %s\n", err)
			return err
		}
		return nil
	},
}

var (
	rootOpts struct {
		enableNotification bool
		cfgFile            string
		apiPort            int
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVar(&rootOpts.cfgFile, "config", "", "config file (default is $HOME/.document-feed-embedder.yaml)")
	rootCmd.PersistentFlags().BoolVar(&rootOpts.enableNotification, "enable-notification", false, "Enable notification after execution")
	rootCmd.PersistentFlags().IntVar(&rootOpts.apiPort, "port", 8080, "API port")
	_ = viper.BindPFlag(config.ApiPortProp.Key, rootCmd.PersistentFlags().Lookup("port"))
}
