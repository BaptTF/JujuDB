package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search QUERY",
	Short: "Search items by text",
	Long:  "Full-text search across item names, descriptions, locations, categories, and notes.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		q := url.Values{}
		q.Set("q", args[0])

		if v, _ := cmd.Flags().GetUint("location-id"); v > 0 {
			q.Set("location_id", strconv.FormatUint(uint64(v), 10))
		}
		if v, _ := cmd.Flags().GetUint("sub-location-id"); v > 0 {
			q.Set("sub_location_id", strconv.FormatUint(uint64(v), 10))
		}
		if v, _ := cmd.Flags().GetUint("category-id"); v > 0 {
			q.Set("category_id", strconv.FormatUint(uint64(v), 10))
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", strconv.Itoa(v))
		}

		data, err := c.Get("/api/search", q)
		if err != nil {
			return err
		}

		var items []itemDTO
		if err := json.Unmarshal(data, &items); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(items) == 0 {
			fmt.Printf("No results for \"%s\".\n", args[0])
			return nil
		}

		fmt.Printf("Search results for \"%s\" (%d results):\n\n", args[0], len(items))
		for _, item := range items {
			fmt.Print(formatItem(item))
			fmt.Println()
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().Uint("location-id", 0, "Filter by location ID")
	searchCmd.Flags().Uint("sub-location-id", 0, "Filter by sub-location ID")
	searchCmd.Flags().Uint("category-id", 0, "Filter by category ID")
	searchCmd.Flags().Int("limit", 0, "Limit number of results")
	rootCmd.AddCommand(searchCmd)
}
