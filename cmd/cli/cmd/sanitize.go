package cmd

import (
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"

	"github.com/spf13/cobra"
)

// sanitizeCmd represents the sanitize command
var sanitizeCmd = &cobra.Command{
	Use:   "sanitize",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewFeedAdapterFromConfigs()
		if err != nil {
			fmt.Printf("failed to create adapter: %s\n", err)
			return err
		}
		if err := a.SanitizeArticlesDB(cmd.Context()); err != nil {
			fmt.Printf("failed to sanitize articles db: %s\n", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sanitizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sanitizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sanitizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
