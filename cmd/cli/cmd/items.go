package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

type itemDTO struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	LocationID    *uint   `json:"location_id"`
	SubLocationID *uint   `json:"sub_location_id"`
	CategoryID    *uint   `json:"category_id"`
	Quantity      int     `json:"quantity"`
	ExpiryDate    *string `json:"expiry_date"`
	AddedDate     string  `json:"added_date"`
	AddedAt       string  `json:"added_at"`
	Notes         *string `json:"notes"`
	Location      string  `json:"location"`
	SubLocation   string  `json:"sub_location"`
	Category      string  `json:"category"`
}

func formatItem(item itemDTO) string {
	s := fmt.Sprintf("  #%d %s\n", item.ID, item.Name)
	if item.Description != "" {
		s += fmt.Sprintf("     Description: %s\n", item.Description)
	}
	loc := item.Location
	if loc == "" {
		loc = "none"
	}
	if item.SubLocation != "" {
		loc += " > " + item.SubLocation
	}
	s += fmt.Sprintf("     Location: %s\n", loc)
	cat := item.Category
	if cat == "" {
		cat = "none"
	}
	s += fmt.Sprintf("     Category: %s\n", cat)
	s += fmt.Sprintf("     Quantity: %d\n", item.Quantity)
	expiry := "none"
	if item.ExpiryDate != nil && *item.ExpiryDate != "" {
		expiry = *item.ExpiryDate
	}
	s += fmt.Sprintf("     Expiry: %s\n", expiry)
	if item.Notes != nil && *item.Notes != "" {
		s += fmt.Sprintf("     Notes: %s\n", *item.Notes)
	}
	return s
}

var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Manage items in the inventory",
}

var itemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List items",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		q := url.Values{}
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
		if v, _ := cmd.Flags().GetInt("offset"); v > 0 {
			q.Set("offset", strconv.Itoa(v))
		}

		data, err := c.Get("/api/items", q)
		if err != nil {
			return err
		}

		var items []itemDTO
		if err := json.Unmarshal(data, &items); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("No items found.")
			return nil
		}

		fmt.Printf("Items (%d total):\n\n", len(items))
		for _, item := range items {
			fmt.Print(formatItem(item))
			fmt.Println()
		}
		return nil
	},
}

var itemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new item",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]interface{}{
			"name": name,
		}

		if v, _ := cmd.Flags().GetInt("quantity"); v > 0 {
			body["quantity"] = v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body["description"] = v
		}
		if v, _ := cmd.Flags().GetUint("location-id"); v > 0 {
			body["location_id"] = v
		}
		if v, _ := cmd.Flags().GetUint("sub-location-id"); v > 0 {
			body["sub_location_id"] = v
		}
		if v, _ := cmd.Flags().GetUint("category-id"); v > 0 {
			body["category_id"] = v
		}
		if v, _ := cmd.Flags().GetString("expiry"); v != "" {
			body["expiry_date"] = v
		}
		if v, _ := cmd.Flags().GetString("notes"); v != "" {
			body["notes"] = v
		}

		jsonBody, _ := json.Marshal(body)
		data, err := c.Post("/api/items", jsonBody)
		if err != nil {
			return err
		}

		var item itemDTO
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Item created:")
		fmt.Println()
		fmt.Print(formatItem(item))
		return nil
	},
}

var itemsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update an existing item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		id := args[0]

		body := map[string]interface{}{}

		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("quantity") {
			v, _ := cmd.Flags().GetInt("quantity")
			body["quantity"] = v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}
		if cmd.Flags().Changed("location-id") {
			v, _ := cmd.Flags().GetUint("location-id")
			body["location_id"] = v
		}
		if cmd.Flags().Changed("sub-location-id") {
			v, _ := cmd.Flags().GetUint("sub-location-id")
			body["sub_location_id"] = v
		}
		if cmd.Flags().Changed("category-id") {
			v, _ := cmd.Flags().GetUint("category-id")
			body["category_id"] = v
		}
		if cmd.Flags().Changed("expiry") {
			v, _ := cmd.Flags().GetString("expiry")
			body["expiry_date"] = v
		}
		if cmd.Flags().Changed("notes") {
			v, _ := cmd.Flags().GetString("notes")
			body["notes"] = v
		}

		if len(body) == 0 {
			return fmt.Errorf("no fields to update. Use flags like --name, --quantity, etc.")
		}

		jsonBody, _ := json.Marshal(body)
		data, err := c.Put("/api/items/"+id, jsonBody)
		if err != nil {
			return err
		}

		var item itemDTO
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Item updated:")
		fmt.Println()
		fmt.Print(formatItem(item))
		return nil
	},
}

var itemsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete an item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		if err := c.Delete("/api/items/"+args[0], nil); err != nil {
			return err
		}

		fmt.Printf("Item #%s deleted.\n", args[0])
		return nil
	},
}

func init() {
	// items list flags
	itemsListCmd.Flags().Uint("location-id", 0, "Filter by location ID")
	itemsListCmd.Flags().Uint("sub-location-id", 0, "Filter by sub-location ID")
	itemsListCmd.Flags().Uint("category-id", 0, "Filter by category ID")
	itemsListCmd.Flags().Int("limit", 0, "Limit number of results")
	itemsListCmd.Flags().Int("offset", 0, "Offset for pagination")

	// items create flags
	itemsCreateCmd.Flags().String("name", "", "Item name (required)")
	itemsCreateCmd.Flags().Int("quantity", 1, "Quantity")
	itemsCreateCmd.Flags().String("description", "", "Item description")
	itemsCreateCmd.Flags().Uint("location-id", 0, "Location ID")
	itemsCreateCmd.Flags().Uint("sub-location-id", 0, "Sub-location ID")
	itemsCreateCmd.Flags().Uint("category-id", 0, "Category ID")
	itemsCreateCmd.Flags().String("expiry", "", "Expiry date (YYYY-MM-DD)")
	itemsCreateCmd.Flags().String("notes", "", "Notes")

	// items update flags
	itemsUpdateCmd.Flags().String("name", "", "Item name")
	itemsUpdateCmd.Flags().Int("quantity", 0, "Quantity")
	itemsUpdateCmd.Flags().String("description", "", "Item description")
	itemsUpdateCmd.Flags().Uint("location-id", 0, "Location ID")
	itemsUpdateCmd.Flags().Uint("sub-location-id", 0, "Sub-location ID")
	itemsUpdateCmd.Flags().Uint("category-id", 0, "Category ID")
	itemsUpdateCmd.Flags().String("expiry", "", "Expiry date (YYYY-MM-DD)")
	itemsUpdateCmd.Flags().String("notes", "", "Notes")

	itemsCmd.AddCommand(itemsListCmd)
	itemsCmd.AddCommand(itemsCreateCmd)
	itemsCmd.AddCommand(itemsUpdateCmd)
	itemsCmd.AddCommand(itemsDeleteCmd)
	rootCmd.AddCommand(itemsCmd)
}
