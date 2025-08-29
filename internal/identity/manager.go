package identity

import (
	"fmt"
)

// Identity represents a git identity configuration
type Identity struct {
	Name    string   // Identity name (work, personal)
	GitName string   // Git user.name
	Email   string   // Git user.email
	Paths   []string // Directory paths
}

// ConfigManager interface to avoid circular imports
type ConfigManager interface {
	AddIncludeIf(identity *Identity) error
	RemoveIncludeIf(name string) error
}

// Manager handles identity operations
type Manager struct {
	identities    map[string]*Identity
	configManager ConfigManager
}

// NewManager creates a new identity manager
func NewManager(configManager ConfigManager) *Manager {
	return &Manager{
		identities:    make(map[string]*Identity),
		configManager: configManager,
	}
}

// LoadIdentities populates the manager with existing identities from a map.
// This is used for initializing the manager at startup from loaded config files
// and does not trigger any write operations.
func (m *Manager) LoadIdentities(identities map[string]*Identity) {
	m.identities = identities
}

// AddIdentity adds a new identity
func (m *Manager) AddIdentity(name, gitName, email string, paths []string) error {
	if m.identities[name] != nil {
		return fmt.Errorf("identity '%s' already exists", name)
	}

	// Create the identity
	identity := &Identity{
		Name:    name,
		GitName: gitName,
		Email:   email,
		Paths:   paths,
	}

	// Add to config manager first (this handles Git config persistence)
	if m.configManager != nil {
		err := m.configManager.AddIncludeIf(identity)
		if err != nil {
			return fmt.Errorf("failed to update git config: %w", err)
		}
	}

	// Add to in-memory storage only after successful config update
	m.identities[name] = identity

	return nil
}

// RemoveIdentity removes an identity
func (m *Manager) RemoveIdentity(name string) error {
	if m.identities[name] == nil {
		return fmt.Errorf("identity '%s' not found", name)
	}

	// Remove from config manager first (this handles Git config cleanup)
	if m.configManager != nil {
		err := m.configManager.RemoveIncludeIf(name)
		if err != nil {
			return fmt.Errorf("failed to remove from git config: %w", err)
		}
	}

	// Remove from in-memory storage only after successful config removal
	delete(m.identities, name)
	return nil
}

// GetIdentity gets an identity by name
func (m *Manager) GetIdentity(name string) (*Identity, error) {
	identity := m.identities[name]
	if identity == nil {
		return nil, fmt.Errorf("identity '%s' not found", name)
	}
	return identity, nil
}

// ListIdentities returns all identities
func (m *Manager) ListIdentities() map[string]*Identity {
	return m.identities
}
