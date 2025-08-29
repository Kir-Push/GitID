package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kir-Push/GitID/internal/identity"
	"gopkg.in/ini.v1"
)

const (
	GitIDSectionStart = "# GitID Managed Section - Do not edit manually"
	GitIDSectionEnd   = "# End GitID Managed Section"
)

// ConfigManager handles git configuration operations
type ConfigManager struct {
	gitConfigPath string // ~/.gitconfig
	identityDir   string // ~/.gitconfig-gitid-*
}

// NewConfigManager creates a new config manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &ConfigManager{
		gitConfigPath: filepath.Join(homeDir, ".gitconfig"),
		identityDir:   homeDir,
	}, nil
}

// LoadExistingIdentities loads identities from existing git config files
func (c *ConfigManager) LoadExistingIdentities() (map[string]*identity.Identity, error) {
	identities := make(map[string]*identity.Identity)

	content, err := c.readGitConfig()
	if err != nil {
		return identities, nil // Return empty if can't read
	}

	startIndex, endIndex := c.findGitIDSection(content)
	if startIndex == -1 {
		return identities, nil // No GitID section
	}

	// Parse includeIf entries in GitID section
	currentIdentity := ""
	var currentPaths []string

	for i := startIndex + 1; i < endIndex; i++ {
		line := strings.TrimSpace(content[i])

		// Match includeIf line: [includeIf "gitdir:/path/"]
		if strings.HasPrefix(line, "[includeIf \"gitdir:") {
			// Extract path from gitdir
			start := strings.Index(line, "gitdir:") + 7
			end := strings.LastIndex(line, "\"]")
			if start < end {
				path := line[start:end]
				// Remove trailing slash
				if strings.HasSuffix(path, "/") {
					path = path[:len(path)-1]
				}
				currentPaths = append(currentPaths, path)
			}
		}

		// Match path line: path = ~/.gitconfig-gitid-name
		if strings.Contains(line, "path = ") && strings.Contains(line, ".gitconfig-gitid-") {
			// Extract identity name from filename
			parts := strings.Split(line, ".gitconfig-gitid-")
			if len(parts) > 1 {
				currentIdentity = strings.TrimSpace(parts[1])

				// Load identity details from identity file
				if ident, err := c.loadIdentityFile(currentIdentity); err == nil {
					ident.Paths = currentPaths
					identities[currentIdentity] = ident
				}

				// Reset for next identity
				currentPaths = []string{}
			}
		}
	}

	return identities, nil
}

// loadIdentityFile loads an identity from its config file
func (c *ConfigManager) loadIdentityFile(name string) (*identity.Identity, error) {
	identityFile := filepath.Join(c.identityDir, fmt.Sprintf(".gitconfig-gitid-%s", name))

	cfg, err := ini.Load(identityFile)
	if err != nil {
		return nil, err
	}

	userSection := cfg.Section("user")
	gitName := userSection.Key("name").String()
	email := userSection.Key("email").String()

	return &identity.Identity{
		Name:    name,
		GitName: gitName,
		Email:   email,
		Paths:   []string{}, // Will be set by caller
	}, nil
}

// AddIncludeIf adds an includeIf entry to ~/.gitconfig and creates identity file
func (c *ConfigManager) AddIncludeIf(identity *identity.Identity) error {
	// 1. Create identity config file
	if err := c.createIdentityFile(identity); err != nil {
		return fmt.Errorf("failed to create identity file: %w", err)
	}

	// 2. Add includeIf entry to ~/.gitconfig
	if err := c.addIncludeIfEntry(identity); err != nil {
		return fmt.Errorf("failed to add includeIf entry: %w", err)
	}

	return nil
}

// RemoveIncludeIf removes an includeIf entry from ~/.gitconfig and deletes identity file
func (c *ConfigManager) RemoveIncludeIf(name string) error {
	// 1. Remove includeIf entry from ~/.gitconfig
	if err := c.removeIncludeIfEntry(name); err != nil {
		return fmt.Errorf("failed to remove includeIf entry: %w", err)
	}

	// 2. Delete identity file
	identityFile := filepath.Join(c.identityDir, fmt.Sprintf(".gitconfig-gitid-%s", name))
	if err := os.Remove(identityFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove identity file: %w", err)
	}

	return nil
}

// createIdentityFile creates ~/.gitconfig-gitid-{name} file
func (c *ConfigManager) createIdentityFile(identity *identity.Identity) error {
	identityFile := filepath.Join(c.identityDir, fmt.Sprintf(".gitconfig-gitid-%s", identity.Name))

	cfg := ini.Empty()
	userSection, err := cfg.NewSection("user")
	if err != nil {
		return err
	}

	userSection.NewKey("name", identity.GitName)
	userSection.NewKey("email", identity.Email)

	return cfg.SaveTo(identityFile)
}

// addIncludeIfEntry adds includeIf entries to ~/.gitconfig
func (c *ConfigManager) addIncludeIfEntry(identity *identity.Identity) error {
	content, err := c.readGitConfig()
	if err != nil {
		return err
	}

	// Find or create GitID managed section
	startIndex, endIndex := c.findGitIDSection(content)

	var newEntries []string
	for _, path := range identity.Paths {
		// Expand ~ to home directory
		expandedPath := path
		if strings.HasPrefix(path, "~/") {
			homeDir, _ := os.UserHomeDir()
			expandedPath = filepath.Join(homeDir, path[2:])
		}

		// Ensure path ends with / for gitdir matching
		if !strings.HasSuffix(expandedPath, "/") {
			expandedPath += "/"
		}

		newEntries = append(newEntries, fmt.Sprintf("[includeIf \"gitdir:%s\"]", expandedPath))
		newEntries = append(newEntries, fmt.Sprintf("    path = %s",
			filepath.Join(c.identityDir, fmt.Sprintf(".gitconfig-gitid-%s", identity.Name))))
	}

	var newContent []string

	if startIndex == -1 {
		// No GitID section exists, add it at the end
		newContent = append(content, "")
		newContent = append(newContent, GitIDSectionStart)
		newContent = append(newContent, newEntries...)
		newContent = append(newContent, GitIDSectionEnd)
	} else {
		// Preserve existing entries and add new ones
		var existingEntries []string
		for i := startIndex + 1; i < endIndex; i++ {
			existingEntries = append(existingEntries, content[i])
		}

		newContent = append(content[:startIndex], GitIDSectionStart)
		newContent = append(newContent, existingEntries...)
		newContent = append(newContent, newEntries...)
		newContent = append(newContent, GitIDSectionEnd)
		if endIndex < len(content) {
			newContent = append(newContent, content[endIndex+1:]...)
		}
	}

	return c.writeGitConfig(newContent)
}

// removeIncludeIfEntry removes includeIf entries for a specific identity
func (c *ConfigManager) removeIncludeIfEntry(name string) error {
	content, err := c.readGitConfig()
	if err != nil {
		return err
	}

	startIndex, endIndex := c.findGitIDSection(content)
	if startIndex == -1 {
		return nil // No GitID section exists
	}

	// Filter out entries for this identity
	var newEntries []string
	identityPath := fmt.Sprintf(".gitconfig-gitid-%s", name)

	i := startIndex + 1
	for i < endIndex {
		line := content[i]
		// Check if this is a path line for the identity we're removing
		if strings.Contains(line, identityPath) {
			// Skip this path line and the preceding includeIf line
			if i > startIndex+1 {
				newEntries = newEntries[:len(newEntries)-1] // Remove the includeIf line
			}
		} else {
			newEntries = append(newEntries, line)
		}
		i++
	}

	var newContent []string
	if len(newEntries) == 0 {
		// Remove entire GitID section if no entries remain
		newContent = append(content[:startIndex], content[endIndex+1:]...)
	} else {
		// Replace with filtered entries
		newContent = append(content[:startIndex], GitIDSectionStart)
		newContent = append(newContent, newEntries...)
		newContent = append(newContent, GitIDSectionEnd)
		if endIndex < len(content) {
			newContent = append(newContent, content[endIndex+1:]...)
		}
	}

	return c.writeGitConfig(newContent)
}

// readGitConfig reads ~/.gitconfig and returns lines
func (c *ConfigManager) readGitConfig() ([]string, error) {
	if _, err := os.Stat(c.gitConfigPath); os.IsNotExist(err) {
		return []string{}, nil // Return empty if file doesn't exist
	}

	data, err := os.ReadFile(c.gitConfigPath)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}

// writeGitConfig writes lines to ~/.gitconfig
func (c *ConfigManager) writeGitConfig(lines []string) error {
	content := strings.Join(lines, "\n")
	return os.WriteFile(c.gitConfigPath, []byte(content), 0644)
}

// findGitIDSection finds the GitID managed section in the config
func (c *ConfigManager) findGitIDSection(lines []string) (int, int) {
	startIndex := -1
	endIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == GitIDSectionStart {
			startIndex = i
		} else if strings.TrimSpace(line) == GitIDSectionEnd {
			endIndex = i
			break
		}
	}

	return startIndex, endIndex
}
