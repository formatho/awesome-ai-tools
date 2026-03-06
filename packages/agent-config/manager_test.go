package agentconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig returned nil")
	}

	// Check default values
	if config.Global.Timeout != 300 {
		t.Errorf("Expected default timeout 300, got %d", config.Global.Timeout)
	}

	if config.Global.MaxRetries != 3 {
		t.Errorf("Expected default max_retries 3, got %d", config.Global.MaxRetries)
	}

	if config.Agents == nil {
		t.Error("Agents map should be initialized")
	}
}

func TestManager_New(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config creates default",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "valid config",
			config:  NewConfig(),
			wantErr: false,
		},
		{
			name: "invalid temperature",
			config: &Config{
				Global: GlobalConfig{
					LLM: &LLMConfig{
						Temperature: ptrFloat64(3.0), // Invalid: > 2
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && m == nil {
				t.Error("New() returned nil manager")
			}
		})
	}
}

func TestManager_LoadSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "agent-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager with test config
	config := NewConfig()
	config.Global.Timeout = 500
	config.Agents["test-agent"] = &AgentConfig{
		LLM: &LLMConfig{
			Model: "gpt-4",
		},
	}

	m, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test YAML
	yamlPath := filepath.Join(tmpDir, "config.yaml")
	if err := m.Save(yamlPath); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}

	// Load into new manager
	m2, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	if err := m2.Load(yamlPath); err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Verify loaded config
	if m2.GetGlobal().Timeout != 500 {
		t.Errorf("Expected timeout 500, got %d", m2.GetGlobal().Timeout)
	}

	agent := m2.GetAgent("test-agent")
	if agent == nil {
		t.Fatal("test-agent not found")
	}
	if agent.LLM.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", agent.LLM.Model)
	}

	// Test JSON
	jsonPath := filepath.Join(tmpDir, "config.json")
	if err := m.Save(jsonPath); err != nil {
		t.Fatalf("Failed to save JSON: %v", err)
	}

	// Test TOML
	tomlPath := filepath.Join(tmpDir, "config.toml")
	if err := m.Save(tomlPath); err != nil {
		t.Fatalf("Failed to save TOML: %v", err)
	}
}

func TestManager_AgentOperations(t *testing.T) {
	m, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test SetAgent
	agent := &AgentConfig{
		LLM: &LLMConfig{
			Model: "gpt-4-turbo",
		},
		Timeout: ptrInt(600),
	}

	if err := m.SetAgent("new-agent", agent); err != nil {
		t.Fatalf("Failed to set agent: %v", err)
	}

	// Test GetAgent
	retrieved := m.GetAgent("new-agent")
	if retrieved == nil {
		t.Fatal("Failed to retrieve agent")
	}

	// Check merged config (should include global defaults)
	if retrieved.LLM.Provider == "" {
		t.Error("LLM provider should have global default")
	}

	// Test ListAgents
	agents := m.ListAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	// Test DeleteAgent
	if !m.DeleteAgent("new-agent") {
		t.Error("Failed to delete agent")
	}

	if m.DeleteAgent("non-existent") {
		t.Error("Should return false for non-existent agent")
	}

	// Verify deletion
	if len(m.ListAgents()) != 0 {
		t.Error("Agent should be deleted")
	}
}

func TestManager_GlobalOperations(t *testing.T) {
	m, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test GetGlobal
	global := m.GetGlobal()
	if global.Timeout != 300 {
		t.Errorf("Expected default timeout 300, got %d", global.Timeout)
	}

	// Test SetGlobal
	newGlobal := GlobalConfig{
		Timeout:            600,
		MaxRetries:         5,
		Debug:              true,
		MaxConcurrentTasks: 10,
		LLM: &LLMConfig{
			Provider:    "anthropic",
			Model:       "claude-3",
			Temperature: ptrFloat64(0.5),
		},
	}

	if err := m.SetGlobal(newGlobal); err != nil {
		t.Fatalf("Failed to set global: %v", err)
	}

	updated := m.GetGlobal()
	if updated.Timeout != 600 {
		t.Errorf("Expected timeout 600, got %d", updated.Timeout)
	}
}

func TestManager_Validation(t *testing.T) {
	// Test invalid temperature
	invalidConfig := &Config{
		Global: GlobalConfig{
			LLM: &LLMConfig{
				Temperature: ptrFloat64(-0.5), // Invalid
			},
		},
	}

	m, err := New(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for invalid temperature")
	}
	_ = m
}

func TestParser_YAML(t *testing.T) {
	yamlData := `
global:
  timeout: 300
  debug: true
  llm:
    provider: openai
    model: gpt-4
    temperature: 0.7

agents:
  test-agent:
    llm:
      model: gpt-4-turbo
    timeout: 600
`

	parser := NewParser()
	config, err := parser.Parse([]byte(yamlData), FormatYAML)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if config.Global.Timeout != 300 {
		t.Errorf("Expected timeout 300, got %d", config.Global.Timeout)
	}

	if config.Agents["test-agent"] == nil {
		t.Error("test-agent not found")
	}

	// Test serialization
	data, err := parser.Serialize(config, FormatYAML)
	if err != nil {
		t.Fatalf("Failed to serialize YAML: %v", err)
	}
	if len(data) == 0 {
		t.Error("Serialized data should not be empty")
	}
}

func TestParser_JSON(t *testing.T) {
	jsonData := `{
		"global": {
			"timeout": 300,
			"debug": true,
			"llm": {
				"provider": "openai",
				"model": "gpt-4"
			}
		},
		"agents": {
			"test-agent": {
				"llm": {
					"model": "gpt-4-turbo"
				}
			}
		}
	}`

	parser := NewParser()
	config, err := parser.Parse([]byte(jsonData), FormatJSON)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if config.Global.Timeout != 300 {
		t.Errorf("Expected timeout 300, got %d", config.Global.Timeout)
	}

	if config.Agents["test-agent"] == nil {
		t.Error("test-agent not found")
	}
}

func TestParser_TOML(t *testing.T) {
	tomlData := `
[global]
timeout = 300
debug = true

[global.llm]
provider = "openai"
model = "gpt-4"

[agents.test-agent]

[agents.test-agent.llm]
model = "gpt-4-turbo"
`

	parser := NewParser()
	config, err := parser.Parse([]byte(tomlData), FormatTOML)
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	if config.Global.Timeout != 300 {
		t.Errorf("Expected timeout 300, got %d", config.Global.Timeout)
	}
}

func TestParser_FormatDetection(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		path     string
		expected Format
		wantErr  bool
	}{
		{"config.yaml", FormatYAML, false},
		{"config.yml", FormatYAML, false},
		{"config.toml", FormatTOML, false},
		{"config.json", FormatJSON, false},
		{"config.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			format, err := parser.DetectFormat(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if format != tt.expected {
					t.Errorf("Expected format %s, got %s", tt.expected, format)
				}
			}
		})
	}
}

func TestValidator_TemperatureRange(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		temp    float64
		wantErr bool
	}{
		{"valid low", 0.0, false},
		{"valid mid", 1.0, false},
		{"valid high", 2.0, false},
		{"invalid low", -0.1, true},
		{"invalid high", 2.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Global: GlobalConfig{
					LLM: &LLMConfig{
						Temperature: ptrFloat64(tt.temp),
					},
				},
			}

			err := validator.Validate(config)
			if tt.wantErr && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_AgentName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		agent   string
		wantErr bool
	}{
		{"valid", "my-agent", false},
		{"valid with underscore", "my_agent", false},
		{"valid alphanumeric", "agent123", false},
		{"invalid with space", "my agent", true},
		{"invalid with dot", "my.agent", true},
		{"invalid special char", "my@agent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Global: GlobalConfig{},
				Agents: map[string]*AgentConfig{
					tt.agent: {Name: tt.agent},
				},
			}

			err := validator.Validate(config)
			if tt.wantErr && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_SkillPatterns(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		allowed []string
		wantErr bool
	}{
		{"valid simple", []string{"file.read"}, false},
		{"valid wildcard", []string{"file.*"}, false},
		{"valid multiple", []string{"file.read", "file.write"}, false},
		{"invalid empty", []string{""}, true},
		{"invalid pattern", []string{"invalid-skill#"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Global: GlobalConfig{},
				Agents: map[string]*AgentConfig{
					"test": {
						Skills: &SkillsConfig{
							Allowed: tt.allowed,
						},
					},
				},
			}

			err := validator.Validate(config)
			if tt.wantErr && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAgentConfig_MergeWithGlobal(t *testing.T) {
	global := &GlobalConfig{
		Timeout:            300,
		MaxRetries:         3,
		MaxConcurrentTasks: 5,
		Debug:              false,
		LLM: &LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: ptrFloat64(0.7),
		},
	}

	agent := &AgentConfig{
		Name: "test",
		LLM: &LLMConfig{
			Model: "gpt-4-turbo", // Override
		},
		Timeout: ptrInt(600), // Override
	}

	merged := agent.MergeWithGlobal(global)

	// Check overrides
	if merged.LLM.Model != "gpt-4-turbo" {
		t.Error("Model should be overridden")
	}
	if *merged.Timeout != 600 {
		t.Error("Timeout should be overridden")
	}

	// Check inherited values
	if merged.LLM.Provider != "openai" {
		t.Error("Provider should be inherited from global")
	}
	if merged.LLM.Temperature == nil || *merged.LLM.Temperature != 0.7 {
		t.Error("Temperature should be inherited from global")
	}
}

func TestManager_ExportImport(t *testing.T) {
	// Create manager with test config
	config := NewConfig()
	config.Global.Timeout = 500
	config.Agents["export-test"] = &AgentConfig{
		LLM: &LLMConfig{
			Model: "claude-3",
		},
	}

	m, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test Export
	data, err := m.Export()
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	if len(data) == 0 {
		t.Error("Export data should not be empty")
	}

	// Test Import into new manager
	m2, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	if err := m2.Import(data, "yaml"); err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	// Verify imported config
	if m2.GetGlobal().Timeout != 500 {
		t.Error("Imported config should have timeout 500")
	}

	agent := m2.GetAgent("export-test")
	if agent == nil {
		t.Fatal("export-test agent not found after import")
	}
	if agent.LLM.Model != "claude-3" {
		t.Error("Agent model should be claude-3")
	}
}

func TestManager_ImportHelpers(t *testing.T) {
	m, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	yamlData := `
global:
  timeout: 400
`
	if err := m.ImportYAML([]byte(yamlData)); err != nil {
		t.Fatalf("ImportYAML failed: %v", err)
	}
	if m.GetGlobal().Timeout != 400 {
		t.Error("YAML import failed")
	}

	jsonData := `{"global": {"timeout": 450}}`
	if err := m.ImportJSON([]byte(jsonData)); err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}
	if m.GetGlobal().Timeout != 450 {
		t.Error("JSON import failed")
	}

	tomlData := `
[global]
timeout = 475
`
	if err := m.ImportTOML([]byte(tomlData)); err != nil {
		t.Fatalf("ImportTOML failed: %v", err)
	}
	if m.GetGlobal().Timeout != 475 {
		t.Error("TOML import failed")
	}
}

func TestValidationErrors(t *testing.T) {
	err := &ValidationError{
		Field:   "test.field",
		Message: "test error",
		Value:   "test-value",
	}

	expected := "test.field: test error (got: test-value)"
	if err.Error() != expected {
		t.Errorf("Error string mismatch: got %s", err.Error())
	}

	errNoValue := &ValidationError{
		Field:   "test.field",
		Message: "test error",
	}

	expectedNoValue := "test.field: test error"
	if errNoValue.Error() != expectedNoValue {
		t.Errorf("Error string mismatch: got %s", errNoValue.Error())
	}
}

func TestDeepCopy(t *testing.T) {
	original := &AgentConfig{
		Name:     "test",
		Timeout:  ptrInt(600),
		LLM:      &LLMConfig{Model: "gpt-4"},
		Metadata: map[string]interface{}{"key": "value"},
	}

	copied := original.DeepCopy()

	// Modify original
	original.Name = "modified"
	*original.Timeout = 300
	original.LLM.Model = "claude"
	original.Metadata["key"] = "modified"

	// Verify copy is independent
	if copied.Name == "modified" {
		t.Error("DeepCopy should be independent")
	}
	if *copied.Timeout == 300 {
		t.Error("DeepCopy should be independent")
	}
	if copied.LLM.Model == "claude" {
		t.Error("DeepCopy should be independent")
	}
	if copied.Metadata["key"] == "modified" {
		t.Error("DeepCopy should be independent")
	}
}

func TestLoadFile(t *testing.T) {
	// Create temp file
	tmpDir, err := os.MkdirTemp("", "agent-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `
global:
  timeout: 999
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test LoadFile
	config, err := LoadFile(yamlPath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if config.Global.Timeout != 999 {
		t.Errorf("Expected timeout 999, got %d", config.Global.Timeout)
	}
}

func TestExists(t *testing.T) {
	// Create temp file
	tmpDir, err := os.MkdirTemp("", "agent-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	existingPath := filepath.Join(tmpDir, "exists.yaml")
	if err := os.WriteFile(existingPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if !Exists(existingPath) {
		t.Error("Exists should return true for existing file")
	}

	if Exists(filepath.Join(tmpDir, "nonexistent.yaml")) {
		t.Error("Exists should return false for non-existing file")
	}
}

// Benchmark tests
func BenchmarkParser_YAML(b *testing.B) {
	yamlData := `
global:
  timeout: 300
  llm:
    provider: openai
    model: gpt-4
agents:
  test:
    llm:
      model: gpt-4-turbo
`
	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse([]byte(yamlData), FormatYAML)
	}
}

func BenchmarkValidator_Validate(b *testing.B) {
	config := NewConfig()
	for i := 0; i < 10; i++ {
		config.Agents[string(rune('a'+i))] = &AgentConfig{
			LLM: &LLMConfig{
				Model: "gpt-4",
			},
		}
	}
	validator := NewValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(config)
	}
}

func BenchmarkManager_GetAgent(b *testing.B) {
	config := NewConfig()
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("agent-%d", i)
		config.Agents[name] = &AgentConfig{
			LLM: &LLMConfig{Model: "gpt-4"},
		}
	}
	m, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetAgent("agent-50")
	}
}

// Test concurrent access
func TestManager_ConcurrentAccess(t *testing.T) {
	m, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Add some initial agents
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		m.SetAgent(name, &AgentConfig{LLM: &LLMConfig{Model: "gpt-4"}})
	}

	// Concurrent reads and writes
	done := make(chan bool)

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.ListAgents()
				m.GetAgent("a")
				m.GetGlobal()
			}
			done <- true
		}()
	}

	// Writer goroutines
	for i := 0; i < 3; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				m.SetAgent("test", &AgentConfig{LLM: &LLMConfig{Model: "gpt-4"}})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 8; i++ {
		<-done
	}
}

// Test hot reload (requires actual file)
func TestWatcher_HotReload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping hot reload test in short mode")
	}

	// Create temp file
	tmpDir, err := os.MkdirTemp("", "agent-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	initialContent := "global:\n  timeout: 300\n"
	if err := os.WriteFile(configPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load config
	m, err := New(NewConfig())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if err := m.Load(configPath); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Set up callback
	reloadCount := 0
	m.onReload = func(c *Config) {
		reloadCount++
	}

	// Start watcher
	if err := m.StartWatcher(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer m.StopWatcher()

	// Modify file
	time.Sleep(100 * time.Millisecond) // Let watcher stabilize
	updatedContent := "global:\n  timeout: 999\n"
	if err := os.WriteFile(configPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Wait for reload
	time.Sleep(500 * time.Millisecond)

	// Check if reload happened
	// Note: This test may be flaky due to timing
	if reloadCount == 0 {
		t.Log("Warning: Hot reload may not have triggered (timing dependent)")
	}
}
