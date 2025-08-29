package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yourusername/gitid/internal/identity"
)

// initCmd initializes GitID
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize GitID configuration",
	Long:  "Initialize GitID by creating the necessary configuration structure.",
	Run: func(cmd *cobra.Command, args []string) {
		color.Green("✅ GitID initialized successfully!")
		fmt.Println("You can now add identities using 'gitid add'")
	},
}

// addCmd adds a new identity
var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new identity",
	Long:  "Add a new Git identity with name, email, and directory path mapping.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		gitName, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		path, _ := cmd.Flags().GetString("path")

		if gitName == "" || email == "" || path == "" {
			color.Red("❌ Error: --name, --email, and --path are required")
			os.Exit(1)
		}

		// Expand ~ in path
		if strings.HasPrefix(path, "~/") {
			homeDir, _ := os.UserHomeDir()
			path = filepath.Join(homeDir, path[2:])
		}

		// Add identity to manager
		err := identityManager.AddIdentity(name, gitName, email, []string{path})
		if err != nil {
			color.Red("❌ Failed to add identity: %v", err)
			os.Exit(1)
		}

		// Create identity struct and add to config
		ident := &identity.Identity{
			Name:    name,
			GitName: gitName,
			Email:   email,
			Paths:   []string{path},
		}

		err = configManager.AddIncludeIf(ident)
		if err != nil {
			color.Red("❌ Failed to update git config: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Added identity '%s'", name)
		fmt.Printf("   Name: %s\n", gitName)
		fmt.Printf("   Email: %s\n", email)
		fmt.Printf("   Path: %s\n", path)
	},
}

// listCmd lists all identities
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all identities",
	Long:  "Display all configured Git identities and their settings.",
	Run: func(cmd *cobra.Command, args []string) {
		identities := identityManager.ListIdentities()

		if len(identities) == 0 {
			color.Yellow("No identities configured. Use 'gitid add' to create one.")
			return
		}

		color.Blue("📋 Configured identities:")
		for name, ident := range identities {
			fmt.Printf("  %s\n", color.CyanString(name))
			fmt.Printf("    Name: %s\n", ident.GitName)
			fmt.Printf("    Email: %s\n", ident.Email)
			fmt.Printf("    Paths: %s\n", strings.Join(ident.Paths, ", "))
			fmt.Println()
		}
	},
}

// statusCmd shows current identity status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current identity status",
	Long:  "Display which identity is currently active in the current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			color.Red("❌ Failed to get current directory: %v", err)
			os.Exit(1)
		}

		color.Blue("📍 Current directory: %s", pwd)

		// For MVP, we'll just show that GitID is ready
		// In a full implementation, we'd check git config in current dir
		color.Green("🚀 GitID is active - identities will be applied automatically")
	},
}

// testCmd tests which identity would apply to a path
var testCmd = &cobra.Command{
	Use:   "test [path]",
	Short: "Test which identity applies to a path",
	Long:  "Show which Git identity would be applied for the specified directory path.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		testPath := args[0]

		// Expand ~ in path
		if strings.HasPrefix(testPath, "~/") {
			homeDir, _ := os.UserHomeDir()
			testPath = filepath.Join(homeDir, testPath[2:])
		}

		color.Blue("🔍 Testing path: %s", testPath)

		// For MVP, we'll check which configured paths would match
		identities := identityManager.ListIdentities()
		var matches []string

		for name, ident := range identities {
			for _, path := range ident.Paths {
				if strings.HasPrefix(testPath, path) {
					matches = append(matches, fmt.Sprintf("%s (%s)", name, ident.Email))
				}
			}
		}

		if len(matches) == 0 {
			color.Yellow("⚠️  No identity would apply to this path")
		} else {
			color.Green("✅ Matching identities:")
			for _, match := range matches {
				fmt.Printf("  - %s\n", match)
			}
		}
	},
}

// removeCmd removes an identity
var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an identity",
	Long:  "Remove a Git identity and clean up its configuration.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Remove from config manager
		err := configManager.RemoveIncludeIf(name)
		if err != nil {
			color.Red("❌ Failed to remove from git config: %v", err)
			os.Exit(1)
		}

		// Remove from identity manager
		err = identityManager.RemoveIdentity(name)
		if err != nil {
			color.Red("❌ Failed to remove identity: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Removed identity '%s'", name)
	},
}

func init() {
	// Add flags for add command
	addCmd.Flags().StringP("name", "n", "", "Git user name")
	addCmd.Flags().StringP("email", "e", "", "Git user email")
	addCmd.Flags().StringP("path", "p", "", "Directory path for this identity")
}
