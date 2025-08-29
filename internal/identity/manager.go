package identity

import "fmt"

// Identity represents a git identity configuration
type Identity struct {
	Name    string   // Identity name (work, personal)
	GitName string   // Git user.name
	Email   string   // Git user.email
	Paths   []string // Directory paths
}

// Manager handles identity operations
type Manager struct {
	identities map[string]*Identity
}

// NewManager creates a new identity manager
func NewManager() *Manager {
	return &Manager{
		identities: make(map[string]*Identity),
	}
}

// AddIdentity adds a new identity
func (m *Manager) AddIdentity(name, gitName, email string, paths []string) error {
	if m.identities[name] != nil {
		return fmt.Errorf("identity '%s' already exists", name)
	}

	m.identities[name] = &Identity{
		Name:    name,
		GitName: gitName,
		Email:   email,
		Paths:   paths,
	}

	return nil
}

// RemoveIdentity removes an identity
func (m *Manager) RemoveIdentity(name string) error {
	if m.identities[name] == nil {
		return fmt.Errorf("identity '%s' not found", name)
	}

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
