package config

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultWorkers      = 1
	defaultMaxAttempts  = 1
	defaultQueueType    = "inmemory"
	defaultLeaseTimeout = 30 * time.Second
)

type Config struct {
	Run     RunConfig             `json:"run" yaml:"run"`
	Queue   QueueConfig           `json:"queue" yaml:"queue"`
	LLM     LLMConfig             `json:"llm" yaml:"llm"`
	Roles   map[string]RoleConfig `json:"roles" yaml:"roles"`
	Observe ObserveConfig         `json:"observe" yaml:"observe"`
}

type RunConfig struct {
	ID          string         `json:"id" yaml:"id"`
	Workers     int            `json:"workers" yaml:"workers"`
	MaxAttempts int            `json:"max_attempts" yaml:"max_attempts"`
	RoleLimits  map[string]int `json:"role_limits" yaml:"role_limits"`
}

type QueueConfig struct {
	Type         string   `json:"type" yaml:"type"`
	LeaseTimeout Duration `json:"lease_timeout" yaml:"lease_timeout"`
}

type LLMConfig struct {
	DefaultProvider string                    `json:"default_provider" yaml:"default_provider"`
	Providers       map[string]ProviderConfig `json:"providers" yaml:"providers"`
}

type ProviderConfig struct {
	Provider        string `json:"provider" yaml:"provider"`
	Model           string `json:"model" yaml:"model"`
	BaseURL         string `json:"base_url" yaml:"base_url"`
	AuthMode        string `json:"auth_mode" yaml:"auth_mode"`
	OAuthProvider   string `json:"oauth_provider" yaml:"oauth_provider"`
	WorkDir         string `json:"work_dir" yaml:"work_dir"`
	APIKeyEnv       string `json:"api_key_env" yaml:"api_key_env"`
	MaxTokens       int    `json:"max_tokens" yaml:"max_tokens"`
	ReasoningEffort string `json:"reasoning_effort" yaml:"reasoning_effort"`
	ThinkingBudget  int    `json:"thinking_budget" yaml:"thinking_budget"`
	ServiceTier     string `json:"service_tier" yaml:"service_tier"`
}

type RoleConfig struct {
	LLMProvider   string `json:"llm_provider" yaml:"llm_provider"`
	MaxIterations int    `json:"max_iterations" yaml:"max_iterations"`
}

type ObserveConfig struct {
	EventsJSONL string `json:"events_jsonl" yaml:"events_jsonl"`
	ReportJSON  string `json:"report_json" yaml:"report_json"`
	HTMLReport  string `json:"html_report" yaml:"html_report"`
}

type RunOptions struct {
	RunID        string
	Workers      int
	MaxAttempts  int
	RoleLimits   map[string]int
	QueueType    string
	LeaseTimeout time.Duration
}

type Duration struct {
	time.Duration
}

func (c *Config) applyDefaults() {
	if c.Run.Workers == 0 {
		c.Run.Workers = defaultWorkers
	}
	if c.Run.MaxAttempts == 0 {
		c.Run.MaxAttempts = defaultMaxAttempts
	}
	if c.Queue.Type == "" {
		c.Queue.Type = defaultQueueType
	}
	if c.Queue.LeaseTimeout.Duration == 0 {
		c.Queue.LeaseTimeout.Duration = defaultLeaseTimeout
	}
}

func (c Config) RunOptions() RunOptions {
	return RunOptions{
		RunID:        c.Run.ID,
		Workers:      c.Run.Workers,
		MaxAttempts:  c.Run.MaxAttempts,
		RoleLimits:   cloneIntMap(c.Run.RoleLimits),
		QueueType:    c.Queue.Type,
		LeaseTimeout: c.Queue.LeaseTimeout.Duration,
	}
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Value == "" {
		d.Duration = 0
		return nil
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		if text == "" {
			d.Duration = 0
			return nil
		}
		parsed, err := time.ParseDuration(text)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", text, err)
		}
		d.Duration = parsed
		return nil
	}

	var nanos int64
	if err := json.Unmarshal(data, &nanos); err != nil {
		return err
	}
	d.Duration = time.Duration(nanos)
	return nil
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
