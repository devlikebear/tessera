package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrInvalidConfig = errors.New("invalid config")

func (c Config) Validate() error {
	var problems []string
	if c.Run.Workers < 1 {
		problems = append(problems, "run.workers must be at least 1")
	}
	if c.Run.MaxAttempts < 1 {
		problems = append(problems, "run.max_attempts must be at least 1")
	}
	for _, role := range sortedKeys(c.Run.RoleLimits) {
		if c.Run.RoleLimits[role] < 1 {
			problems = append(problems, fmt.Sprintf("run.role_limits.%s must be at least 1", role))
		}
	}
	if c.Queue.Type != "inmemory" {
		problems = append(problems, "queue.type must be inmemory")
	}
	if c.Queue.LeaseTimeout.Duration <= 0 {
		problems = append(problems, "queue.lease_timeout must be positive")
	}
	if c.LLM.DefaultProvider != "" && !providerExists(c.LLM.Providers, c.LLM.DefaultProvider) {
		problems = append(problems, fmt.Sprintf("llm.default_provider %q is not defined in llm.providers", c.LLM.DefaultProvider))
	}
	for _, role := range sortedKeys(c.Roles) {
		provider := c.Roles[role].LLMProvider
		if provider == "" {
			continue
		}
		if !providerExists(c.LLM.Providers, provider) {
			problems = append(problems, fmt.Sprintf("roles.%s.llm_provider %q is not defined in llm.providers", role, provider))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidConfig, strings.Join(problems, "; "))
	}
	return nil
}

func providerExists(providers map[string]ProviderConfig, name string) bool {
	if providers == nil {
		return false
	}
	_, ok := providers[name]
	return ok
}

func sortedKeys[V any](in map[string]V) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
