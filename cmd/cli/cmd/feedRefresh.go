package cmd

import (
	"fmt"
	"github.com/eldius/document-feeder/internal/ui"

	"github.com/eldius/document-feeder/internal/adapter"

	"github.com/spf13/cobra"
)

// feedRefreshCmd represents the refresh command.
var feedRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			err := fmt.Errorf("creating adapter: %w", err)
			fmt.Printf("failed to create adapter: %s\n", err)
			return err
		}
		if cmdRefreshOpts.interactive {
			a, err := adapter.NewDefaultAdapter()
			if err != nil {
				err := fmt.Errorf("creating adapter: %w", err)
				fmt.Printf("failed to create adapter: %s\n", err)
				return err
			}
			if err := ui.RefreshScreen(cmd.Context(), a); err != nil {
				fmt.Printf("failed to refresh screen: %s\n", err)
				return err
			}
			return nil
		}
		fmt.Println("refreshing feeds")
		if err := a.Refresh(cmd.Context()); err != nil {
			err := fmt.Errorf("refreshing feeds: %w", err)
			fmt.Printf("failed to refresh feeds: %s\n", err)
			return err
		}
		return nil
	},
}

var (
	cmdRefreshOpts struct {
		interactive bool
	}
)

func init() {
	feedCmd.AddCommand(feedRefreshCmd)

	feedRefreshCmd.Flags().BoolVarP(&cmdRefreshOpts.interactive, "interactive", "i", false, "Visual execution feedback for the user.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
