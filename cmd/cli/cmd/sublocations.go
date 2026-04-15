package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

type subLocation struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	LocationID uint   `json:"location_id"`
}

func formatSubLocation(sl subLocation) string {
	return fmt.Sprintf("  #%d %s (location_id: %d)\n", sl.ID, sl.Name, sl.LocationID)
}

var sublocationsCmd = &cobra.Command{
	Use:   "sublocations",
	Short: "Manage sub-locations",
}

var sublocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sub-locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		q := url.Values{}
		if v, _ := cmd.Flags().GetUint("location-id"); v > 0 {
			q.Set("location_id", strconv.FormatUint(uint64(v), 10))
		}

		data, err := c.Get("/api/sub-locations", q)
		if err != nil {
			return err
		}

		var subs []subLocation
		if err := json.Unmarshal(data, &subs); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(subs) == 0 {
			fmt.Println("No sub-locations found.")
			return nil
		}

		fmt.Printf("Sub-locations (%d total):\n\n", len(subs))
		for _, sl := range subs {
			fmt.Print(formatSubLocation(sl))
		}
		return nil
	},
}

var sublocationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new sub-location",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		locationID, _ := cmd.Flags().GetUint("location-id")
		if locationID == 0 {
			return fmt.Errorf("--location-id is required")
		}

		body, _ := json.Marshal(map[string]interface{}{
			"name":        name,
			"location_id": locationID,
		})
		data, err := c.Post("/api/sub-locations", body)
		if err != nil {
			return err
		}

		var sl subLocation
		if err := json.Unmarshal(data, &sl); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Sub-location created:")
		fmt.Print(formatSubLocation(sl))
		return nil
	},
}

var sublocationsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a sub-location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("location-id") {
			v, _ := cmd.Flags().GetUint("location-id")
			body["location_id"] = v
		}

		if len(body) == 0 {
			return fmt.Errorf("no fields to update. Use flags like --name, --location-id")
		}

		jsonBody, _ := json.Marshal(body)
		data, err := c.Put("/api/sub-locations/"+args[0], jsonBody)
		if err != nil {
			return err
		}

		var sl subLocation
		if err := json.Unmarshal(data, &sl); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Sub-location updated:")
		fmt.Print(formatSubLocation(sl))
		return nil
	},
}

var sublocationsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a sub-location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		q := url.Values{}
		if force, _ := cmd.Flags().GetBool("force"); force {
			q.Set("force", "true")
		}

		if err := c.Delete("/api/sub-locations/"+args[0], q); err != nil {
			return err
		}

		fmt.Printf("Sub-location #%s deleted.\n", args[0])
		return nil
	},
}

func init() {
	sublocationsListCmd.Flags().Uint("location-id", 0, "Filter by parent location ID")

	sublocationsCreateCmd.Flags().String("name", "", "Sub-location name (required)")
	sublocationsCreateCmd.Flags().Uint("location-id", 0, "Parent location ID (required)")

	sublocationsUpdateCmd.Flags().String("name", "", "Sub-location name")
	sublocationsUpdateCmd.Flags().Uint("location-id", 0, "Parent location ID")

	sublocationsDeleteCmd.Flags().Bool("force", false, "Force deletion even with dependencies")

	sublocationsCmd.AddCommand(sublocationsListCmd)
	sublocationsCmd.AddCommand(sublocationsCreateCmd)
	sublocationsCmd.AddCommand(sublocationsUpdateCmd)
	sublocationsCmd.AddCommand(sublocationsDeleteCmd)
	rootCmd.AddCommand(sublocationsCmd)
}
