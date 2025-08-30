package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// expandPath expands a path that starts with ~/ to the user's home directory.
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, path[2:]), nil
}

// findMatchingIdentities finds identities that match a given path.
func findMatchingIdentities(testPath string) []string {
	identities := identityManager.ListIdentities()
	var matches []string

	for name, ident := range identities {
		for _, path := range ident.Paths {
			if testPath == path || strings.HasPrefix(testPath, path+string(os.PathSeparator)) {
				matches = append(matches, fmt.Sprintf("%s (%s)", name, ident.Email))
			}
		}
	}
	return matches
}

func printMatchesForPath(path string) {
	matches := findMatchingIdentities(path)

	if len(matches) == 0 {
		color.Yellow("⚠️  No identity would apply to this path")
	} else {
		color.Green("✅ Matching identities:")
		for _, match := range matches {
			fmt.Printf("  - %s\n", match)
		}
	}
}

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
		paths, _ := cmd.Flags().GetStringArray("path")
		expandedPaths := paths

		if gitName == "" || email == "" || len(paths) == 0 {
			color.Red("❌ Error: --name, --email, and --path are required")
			os.Exit(1)
		}

		for i, path := range paths {
			// Expand ~ in path
			expandedPath, err := expandPath(path)
			if err != nil {
				color.Red("❌ Error expanding path: %v", err)
				os.Exit(1)
			}
			expandedPaths[i] = expandedPath
		}

		// Add identity (this handles both in-memory and config operations)
		err := identityManager.AddIdentity(name, gitName, email, expandedPaths)
		if err != nil {
			color.Red("❌ Failed to add identity: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Added identity '%s'", name)
		fmt.Printf("   Name: %s\n", gitName)
		fmt.Printf("   Email: %s\n", email)
		fmt.Printf("   Paths: %s\n", expandedPaths)
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

		printMatchesForPath(pwd)
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
		expandedPath, err := expandPath(testPath)
		if err != nil {
			color.Red("❌ Error expanding path: %v", err)
			os.Exit(1)
		}
		testPath = expandedPath

		color.Blue("🔍 Testing path: %s", testPath)

		printMatchesForPath(testPath)
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

		// Remove identity (this handles both config and in-memory removal)
		err := identityManager.RemoveIdentity(name)
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
	addCmd.Flags().StringArrayP("path", "p", []string{}, "Directory path for this identity")
}
