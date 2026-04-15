package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

type category struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func formatCategory(cat category) string {
	return fmt.Sprintf("  #%d %s\n", cat.ID, cat.Name)
}

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage categories",
}

var categoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		data, err := c.Get("/api/categories", nil)
		if err != nil {
			return err
		}

		var categories []category
		if err := json.Unmarshal(data, &categories); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(categories) == 0 {
			fmt.Println("No categories found.")
			return nil
		}

		fmt.Printf("Categories (%d total):\n\n", len(categories))
		for _, cat := range categories {
			fmt.Print(formatCategory(cat))
		}
		return nil
	},
}

var categoriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new category",
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
		data, err := c.Post("/api/categories", body)
		if err != nil {
			return err
		}

		var cat category
		if err := json.Unmarshal(data, &cat); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Category created:")
		fmt.Print(formatCategory(cat))
		return nil
	},
}

var categoriesUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a category",
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
		data, err := c.Put("/api/categories/"+args[0], body)
		if err != nil {
			return err
		}

		var cat category
		if err := json.Unmarshal(data, &cat); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("Category updated:")
		fmt.Print(formatCategory(cat))
		return nil
	},
}

var categoriesDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a category",
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

		if err := c.Delete("/api/categories/"+args[0], q); err != nil {
			return err
		}

		fmt.Printf("Category #%s deleted.\n", args[0])
		return nil
	},
}

func init() {
	categoriesCreateCmd.Flags().String("name", "", "Category name (required)")
	categoriesUpdateCmd.Flags().String("name", "", "Category name (required)")
	categoriesDeleteCmd.Flags().Bool("force", false, "Force deletion even with dependencies")

	categoriesCmd.AddCommand(categoriesListCmd)
	categoriesCmd.AddCommand(categoriesCreateCmd)
	categoriesCmd.AddCommand(categoriesUpdateCmd)
	categoriesCmd.AddCommand(categoriesDeleteCmd)
	rootCmd.AddCommand(categoriesCmd)
}
