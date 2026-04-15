package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

type location struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func formatLocation(loc location) string {
	return fmt.Sprintf("  #%d %s\n", loc.ID, loc.Name)
}

var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage locations",
}

var locationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		data, err := c.Get("/api/locations", nil)
		if err != nil {
			return err
		}

		var locations []location
		if err := json.Unmarshal(data, &locations); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(locations) == 0 {
			fmt.Println("No locations found.")
			return nil
		}

		fmt.Printf("Locations (%d total):\n\n", len(locations))
		for _, loc := range locations {
			fmt.Print(formatLocation(loc))
		}
		return nil
	},
}

var locationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new location",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body, _ := json.Marshal(map[string]string{"name": name})
		data, err := c.Post("/api/locations", body)
		if err != nil {
			return err
		}

		var loc location
		if err := json.Unmarshal(data, &loc); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Location created:")
		fmt.Print(formatLocation(loc))
		return nil
	},
}

var locationsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body, _ := json.Marshal(map[string]string{"name": name})
		data, err := c.Put("/api/locations/"+args[0], body)
		if err != nil {
			return err
		}

		var loc location
		if err := json.Unmarshal(data, &loc); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Location updated:")
		fmt.Print(formatLocation(loc))
		return nil
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a location",
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

		if err := c.Delete("/api/locations/"+args[0], q); err != nil {
			return err
		}

		fmt.Printf("Location #%s deleted.\n", args[0])
		return nil
	},
}

func init() {
	locationsCreateCmd.Flags().String("name", "", "Location name (required)")
	locationsUpdateCmd.Flags().String("name", "", "Location name (required)")
	locationsDeleteCmd.Flags().Bool("force", false, "Force deletion even with dependencies")

	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsCreateCmd)
	locationsCmd.AddCommand(locationsUpdateCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	rootCmd.AddCommand(locationsCmd)
}
