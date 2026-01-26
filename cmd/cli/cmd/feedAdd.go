package cmd

import (
	"context"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/adapter"
	"github.com/eldius/document-feed-embedder/internal/model"
	"github.com/eldius/document-feed-embedder/internal/ui"
	"github.com/spf13/cobra"
)

// feedAddCmd represents the add command
var feedAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add RSS feeds to the feed list",
	Long: `Add RSS feeds to the feed list.
Feeds are added using their URL.
Example: feed add https://www.heise.de/news/rss/heise-newsfeed.xml.
`,
	Run: func(cmd *cobra.Command, args []string) {
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			panic(err)
		}

		for _, f := range feedAddOpts.feed {
			feed := processFeed(cmd.Context(), a, f)
			fmt.Printf(" ==> Feed %s added successfully. (%d articles)\n", feed.Title, len(feed.Items))
		}
	},
}

func processFeed(ctx context.Context, a *adapter.FeedAdapter, feedURL string) *model.Feed {
	cancel := ui.ProcessingScreen(ctx, fmt.Sprintf("Processing feed %s...", feedURL))
	defer cancel()

	feed, err := a.Parse(ctx, feedURL)
	if err != nil {
		panic(err)
	}
	return feed
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
