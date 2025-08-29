package cli

import (
	"os"

	"github.com/Kir-Push/GitID/internal/config"
	"github.com/Kir-Push/GitID/internal/identity"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	configManager   *config.ConfigManager
	identityManager *identity.Manager
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitid",
	Short: "Git Identity Manager - Manage Git identities with includeIf",
	Long: `GitID is a lightweight CLI tool that manages Git's includeIf configuration feature,
making it easy to automatically use different Git identities based on directory location.

Examples:
  gitid init                                          # Initialize GitID
  gitid add work --name "John Doe" --email john@company.com --path ~/work
  gitid list                                          # List all identities
  gitid status                                        # Show current identity
  gitid test ~/work/project                          # Test which identity applies`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	// Initialize managers
	var err error
	configManager, err = config.NewConfigManager()
	if err != nil {
		color.Red("Failed to initialize config manager: %v", err)
		os.Exit(1)
	}

	identityManager = identity.NewManager(configManager)

	// Load existing identities from git config
	if existingIdentities, err := configManager.LoadExistingIdentities(); err != nil {
		// Warn about errors but don't exit, so the app remains usable
		// to create a new config or add identities.
		color.Yellow("Warning: could not load existing identities: %v", err)
	} else if len(existingIdentities) > 0 {
		identityManager.LoadIdentities(existingIdentities)
	}

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(removeCmd)
}
