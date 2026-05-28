package sysproxy

// Manager sets and unsets system-level proxy settings.
type Manager struct {
	httpHost string
	httpPort int
}

// NewManager creates a system proxy manager.
func NewManager(httpHost string, httpPort int) *Manager {
	return &Manager{httpHost: httpHost, httpPort: httpPort}
}
