package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/document-feeder/internal/ui"
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
	Run: func(cmd *cobra.Command, args []string) {
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			panic(err)
		}

		errorStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("232")). // Dark text color for contrast
			Background(lipgloss.Color("1")). // Red background
			PaddingLeft(1).
			PaddingRight(1).
			MarginRight(1)

		infoStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")). // Black text color
			Background(lipgloss.Color("6")). // Cyan background
			PaddingLeft(1).
			PaddingRight(1).
			MarginRight(1)

		feedbackMessage := make([]string, len(feedAddOpts.feed))
		for _, f := range feedAddOpts.feed {
			feed := processFeed(cmd.Context(), a, f)
			if feed == nil {
				feedbackMessage = append(feedbackMessage, errorStyle.Render(fmt.Sprintf(" ==> ! Failed to parse %s", f)))
				continue
			}
			feedbackMessage = append(feedbackMessage, infoStyle.Render(fmt.Sprintf(" ==> Feed %s added successfully. (%d articles)\n", feed.Title, len(feed.Items))))
		}

		for _, msg := range feedbackMessage {
			fmt.Print(msg)
		}
	},
}

func processFeed(ctx context.Context, a *adapter.FeedAdapter, feedURL string) *model.Feed {
	cancel := ui.ProcessingScreen(ctx, fmt.Sprintf("Processing feed %s...", feedURL))
	defer cancel()

	feed, err := a.Parse(ctx, feedURL)
	if err != nil {
		return nil
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
