package cmd

import (
	"fmt"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-synchronize the search index",
	Long:  "Re-index all items from the database into MeiliSearch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		_, err = c.PostNoContent("/api/sync/all")
		if err != nil {
			return err
		}

		fmt.Println("Search index synchronized successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
