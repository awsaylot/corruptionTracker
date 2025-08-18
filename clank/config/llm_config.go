package config

// LLMConfig represents configuration for the LLM client
type LLMConfig struct {
	URL      string                 `yaml:"url"`
	Model    string                 `yaml:"model"`
	Timeout  string                 `yaml:"timeout"`
	Options  map[string]interface{} `yaml:"options,omitempty"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}
