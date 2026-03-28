package cmd

import (
	"fmt"
	"github.com/eldius/document-feeder/internal/ui/v2/add_feeds"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/spf13/cobra"
)

// feedAddCmd represents the add command.
var feedAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add RSS feeds to the feed list",
	Long: `Add RSS feeds to the feed list.
Feeds are added using their URL.
Example: feed add https://www.heise.de/news/rss/heise-newsfeed.xml.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewFeedAdapterFromConfigs()
		if err != nil {
			err := fmt.Errorf("creating adapter: %w", err)
			fmt.Printf("failed to create adapter: %s\n", err)
			return err
		}
		//if err := ui.AddScreen(cmd.Context(), a, feedAddOpts.feed); err != nil {
		//	err := fmt.Errorf("adding feeds: %w", err)
		//	fmt.Printf("failed to add feeds: %s\n", err)
		//	return err
		//}
		if err := add_feeds.Start(cmd.Context(), a, feedAddOpts.feed); err != nil {
			err := fmt.Errorf("adding feeds: %w", err)
			fmt.Printf("failed to add feeds: %s\n", err)
			return err
		}
		return nil
	},
}

var (
	feedAddOpts struct {
		feed []string
	}
)

func init() {
	feedCmd.AddCommand(feedAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// feedAddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// feedAddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	feedAddCmd.Flags().StringSliceVarP(&feedAddOpts.feed, "feed", "f", []string{}, "feed to add")
}
