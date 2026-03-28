package cmd

import (
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/ui/v2/add_feed"
	"github.com/spf13/cobra"
)

// testingCmd represents the testing command.
var testingCmd = &cobra.Command{
	Use:   "testing",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewFeedAdapterFromConfigs()
		if err != nil {
			cmd.Printf("failed to create adapter: %s\n", err)
			return err
		}

		if err := add_feed.Start(cmd.Context(), a); err != nil {
			cmd.Printf("failed to start add feed: %s\n", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
