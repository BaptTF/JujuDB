package cmd

import (
	"fmt"

	"jujudb/cmd/cli/client"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the JujuDB server",
	Long:  "Login to a JujuDB server. The session is saved locally and persists across commands.",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		password, _ := cmd.Flags().GetString("password")

		if server == "" {
			return fmt.Errorf("--server is required")
		}
		if password == "" {
			return fmt.Errorf("--password is required")
		}

		if err := client.Login(server, password); err != nil {
			return err
		}

		fmt.Printf("Logged in to %s successfully.\n", server)
		return nil
	},
}

func init() {
	loginCmd.Flags().String("server", "", "JujuDB server URL (e.g. https://jujudb.example.com)")
	loginCmd.Flags().String("password", "", "Authentication password")
	rootCmd.AddCommand(loginCmd)
}
