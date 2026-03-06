package agentconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles configuration management with thread-safe operations.
// It supports loading from files, validation, and hot reload capabilities.
type Manager struct {
	mu     sync.RWMutex
	config *Config

	// parser handles multi-format parsing
	parser *Parser

	// validator handles configuration validation
	validator *Validator

	// watcher handles file watching for hot reload
	watcher *Watcher

	// filePath stores the last loaded/saved file path
	filePath string

	// onReload callback when config is reloaded
	onReload func(*Config)
}

// ManagerOption is a functional option for configuring the Manager.
type ManagerOption func(*Manager) error

// WithOnReload sets a callback for when configuration is reloaded.
func WithOnReload(callback func(*Config)) ManagerOption {
	return func(m *Manager) error {
		m.onReload = callback
		return nil
	}
}

// WithHotReload enables hot reload for the specified file.
func WithHotReload(path string) ManagerOption {
	return func(m *Manager) error {
		m.filePath = path
		return nil
	}
}

// WithValidationRules adds custom validation rules.
func WithValidationRules(rules ...ValidationRule) ManagerOption {
	return func(m *Manager) error {
		for _, rule := range rules {
			m.validator.AddRule(rule)
		}
		return nil
	}
}

// New creates a new Manager with the provided configuration.
func New(config *Config, opts ...ManagerOption) (*Manager, error) {
	if config == nil {
		config = NewConfig()
	}

	m := &Manager{
		config:    config,
		parser:    NewParser(),
		validator: NewValidator(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate initial config
	if err := m.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return m, nil
}

// Load loads configuration from a file, auto-detecting the format.
// The file extension determines the format (.yaml, .yml, .toml, .json).
func (m *Manager) Load(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", path)
	}

	// Parse the file
	config, err := m.parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate the loaded config
	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config
	m.filePath = path

	return nil
}

// Save saves the current configuration to a file.
// The file extension determines the format.
func (m *Manager) Save(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Detect format and serialize
	format, err := m.parser.DetectFormat(path)
	if err != nil {
		return err
	}

	data, err := m.parser.Serialize(m.config, format)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	m.filePath = path
	return nil
}

// GetGlobal returns a copy of the global configuration.
func (m *Manager) GetGlobal() GlobalConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config.Global
}

// SetGlobal updates the global configuration.
func (m *Manager) SetGlobal(global GlobalConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a temporary config for validation
	tempConfig := &Config{
		Global: global,
		Agents: m.config.Agents,
	}

	if err := m.validator.Validate(tempConfig); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	m.config.Global = global
	return nil
}

// GetAgent returns the configuration for a specific agent.
// Returns nil if the agent doesn't exist.
// The returned config is merged with global defaults.
func (m *Manager) GetAgent(name string) *AgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.config.Agents[name]
	if !exists {
		return nil
	}

	// Return a merged copy with global defaults
	return agent.MergeWithGlobal(&m.config.Global)
}

// GetAgentRaw returns the agent config without global defaults.
func (m *Manager) GetAgentRaw(name string) *AgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.config.Agents[name]
	if !exists {
		return nil
	}

	// Return a copy
	copy := *agent
	return &copy
}

// SetAgent sets or updates an agent configuration.
func (m *Manager) SetAgent(name string, agent *AgentConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	if agent == nil {
		return fmt.Errorf("agent configuration cannot be nil")
	}

	// Set the name
	agent.Name = name

	// Initialize agents map if needed
	if m.config.Agents == nil {
		m.config.Agents = make(map[string]*AgentConfig)
	}

	// Validate this specific agent
	tempConfig := &Config{
		Global: m.config.Global,
		Agents: make(map[string]*AgentConfig),
	}
	tempConfig.Agents[name] = agent

	if err := m.validator.Validate(tempConfig); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	m.config.Agents[name] = agent
	return nil
}

// ListAgents returns a list of all agent names.
func (m *Manager) ListAgents() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.config.Agents))
	for name := range m.config.Agents {
		names = append(names, name)
	}
	return names
}

// DeleteAgent removes an agent from the configuration.
// Returns false if the agent doesn't exist.
func (m *Manager) DeleteAgent(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Agents[name]; !exists {
		return false
	}

	delete(m.config.Agents, name)
	return true
}

// Validate validates the entire configuration.
func (m *Manager) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.validator.Validate(m.config)
}

// Export exports the configuration as YAML bytes.
func (m *Manager) Export() ([]byte, error) {
	return m.ExportAs(FormatYAML)
}

// ExportAs exports the configuration in the specified format.
func (m *Manager) ExportAs(format Format) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.parser.Serialize(m.config, format)
}

// Import imports configuration from raw data in the specified format.
func (m *Manager) Import(data []byte, format string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f := Format(format)
	config, err := m.parser.Parse(data, f)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	if err := m.validator.Validate(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	m.config = config
	return nil
}

// ImportYAML imports configuration from YAML data.
func (m *Manager) ImportYAML(data []byte) error {
	return m.Import(data, string(FormatYAML))
}

// ImportTOML imports configuration from TOML data.
func (m *Manager) ImportTOML(data []byte) error {
	return m.Import(data, string(FormatTOML))
}

// ImportJSON imports configuration from JSON data.
func (m *Manager) ImportJSON(data []byte) error {
	return m.Import(data, string(FormatJSON))
}

// StartWatcher starts watching the loaded file for changes.
// When the file changes, it will be reloaded automatically.
func (m *Manager) StartWatcher() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.filePath == "" {
		return fmt.Errorf("no file path set, load a file first")
	}

	if m.watcher != nil {
		return fmt.Errorf("watcher already running")
	}

	watcher, err := NewWatcher(m.filePath, func() {
		m.reloadFile()
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	m.watcher = watcher
	return nil
}

// StopWatcher stops the file watcher.
func (m *Manager) StopWatcher() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watcher == nil {
		return nil
	}

	err := m.watcher.Close()
	m.watcher = nil
	return err
}

// reloadFile reloads the configuration from the last loaded file.
// This is called by the watcher when the file changes.
func (m *Manager) reloadFile() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.filePath == "" {
		return
	}

	config, err := m.parser.ParseFile(m.filePath)
	if err != nil {
		// Log error but don't change the config
		return
	}

	if err := m.validator.Validate(config); err != nil {
		// Log error but don't change the config
		return
	}

	m.config = config

	// Call the reload callback if set
	if m.onReload != nil {
		go m.onReload(config)
	}
}

// Config returns a deep copy of the entire configuration.
func (m *Manager) Config() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy by serializing and deserializing
	data, err := m.parser.Serialize(m.config, FormatYAML)
	if err != nil {
		return nil
	}

	config, err := m.parser.Parse(data, FormatYAML)
	if err != nil {
		return nil
	}

	return config
}

// FilePath returns the path of the last loaded/saved file.
func (m *Manager) FilePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.filePath
}

// Exists returns true if a configuration file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// LoadFile is a convenience function to load a configuration file.
// It creates a temporary Manager and returns the loaded Config.
func LoadFile(path string) (*Config, error) {
	m, err := New(NewConfig())
	if err != nil {
		return nil, err
	}

	if err := m.Load(path); err != nil {
		return nil, err
	}

	return m.Config(), nil
}

// readFileImpl reads a file from disk.
func readFileImpl(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func init() {
	// Set the readFile function
	readFile = readFileImpl
}

// DeepCopy creates a deep copy of an AgentConfig.
func (a *AgentConfig) DeepCopy() *AgentConfig {
	if a == nil {
		return nil
	}

	copy := &AgentConfig{
		Name:               a.Name,
		Timeout:            a.Timeout,
		MaxRetries:         a.MaxRetries,
		MaxConcurrentTasks: a.MaxConcurrentTasks,
		Debug:              a.Debug,
		Enabled:            a.Enabled,
	}

	if a.LLM != nil {
		copy.LLM = &LLMConfig{
			Provider:         a.LLM.Provider,
			Model:            a.LLM.Model,
			BaseURL:          a.LLM.BaseURL,
			APIKey:           a.LLM.APIKey,
			Temperature:      a.LLM.Temperature,
			MaxTokens:        a.LLM.MaxTokens,
			TopP:             a.LLM.TopP,
			FrequencyPenalty: a.LLM.FrequencyPenalty,
			PresencePenalty:  a.LLM.PresencePenalty,
		}
		if len(a.LLM.StopSequences) > 0 {
			copy.LLM.StopSequences = append([]string{}, a.LLM.StopSequences...)
		}
	}

	if a.Skills != nil {
		copy.Skills = &SkillsConfig{}
		if len(a.Skills.Allowed) > 0 {
			copy.Skills.Allowed = append([]string{}, a.Skills.Allowed...)
		}
		if len(a.Skills.Denied) > 0 {
			copy.Skills.Denied = append([]string{}, a.Skills.Denied...)
		}
	}

	if a.Metadata != nil {
		copy.Metadata = make(map[string]interface{})
		for k, v := range a.Metadata {
			copy.Metadata[k] = v
		}
	}

	// Handle pointer values
	if a.Timeout != nil {
		timeout := *a.Timeout
		copy.Timeout = &timeout
	}
	if a.MaxRetries != nil {
		maxRetries := *a.MaxRetries
		copy.MaxRetries = &maxRetries
	}
	if a.MaxConcurrentTasks != nil {
		maxConcurrentTasks := *a.MaxConcurrentTasks
		copy.MaxConcurrentTasks = &maxConcurrentTasks
	}
	if a.Debug != nil {
		debug := *a.Debug
		copy.Debug = &debug
	}
	if a.Enabled != nil {
		enabled := *a.Enabled
		copy.Enabled = &enabled
	}

	return copy
}

// ExportToBuffer exports the config to a buffer for inspection.
func (m *Manager) ExportToBuffer(format Format) (*bytes.Buffer, error) {
	data, err := m.ExportAs(format)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}
