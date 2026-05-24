package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfiguredNovelTeam(t *testing.T) {
	cfg, err := LoadFile("../../examples/configured-novel-team/tessera.yaml")
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.Run.ID != "novel-team-run" {
		t.Fatalf("run id = %q, want novel-team-run", cfg.Run.ID)
	}
	if cfg.Run.Workers != 3 {
		t.Fatalf("workers = %d, want 3", cfg.Run.Workers)
	}
	if cfg.Run.MaxAttempts != 2 {
		t.Fatalf("max attempts = %d, want 2", cfg.Run.MaxAttempts)
	}
	if cfg.Queue.Type != "inmemory" {
		t.Fatalf("queue type = %q, want inmemory", cfg.Queue.Type)
	}
	if cfg.Queue.LeaseTimeout.Duration != 30*time.Second {
		t.Fatalf("lease timeout = %s, want 30s", cfg.Queue.LeaseTimeout.Duration)
	}
	if cfg.LLM.DefaultProvider != "claude" {
		t.Fatalf("default provider = %q, want claude", cfg.LLM.DefaultProvider)
	}
	claude := cfg.LLM.Providers["claude"]
	if claude.Provider != "anthropic" || claude.Model != "claude-sonnet-4" {
		t.Fatalf("claude provider = %+v, want anthropic claude-sonnet-4", claude)
	}
	if claude.APIKeyEnv != "ANTHROPIC_API_KEY" {
		t.Fatalf("api key env = %q, want ANTHROPIC_API_KEY", claude.APIKeyEnv)
	}
	if cfg.Roles["writer"].LLMProvider != "codex" {
		t.Fatalf("writer llm provider = %q, want codex", cfg.Roles["writer"].LLMProvider)
	}
	if cfg.Roles["writer"].MaxIterations != 4 {
		t.Fatalf("writer max iterations = %d, want 4", cfg.Roles["writer"].MaxIterations)
	}

	opts := cfg.RunOptions()
	if opts.RunID != "novel-team-run" {
		t.Fatalf("RunOptions RunID = %q, want novel-team-run", opts.RunID)
	}
	if opts.Workers != 3 || opts.MaxAttempts != 2 {
		t.Fatalf("RunOptions = %+v, want workers=3 maxAttempts=2", opts)
	}
	if opts.RoleLimits["writer"] != 2 {
		t.Fatalf("writer role limit = %d, want 2", opts.RoleLimits["writer"])
	}
	if opts.LeaseTimeout != 30*time.Second {
		t.Fatalf("RunOptions lease timeout = %s, want 30s", opts.LeaseTimeout)
	}
}

func TestLoadJSONAppliesDefaults(t *testing.T) {
	path := writeTempConfig(t, "tessera.json", `{
		"llm": {
			"default_provider": "stub",
			"providers": {
				"stub": {"provider": "anthropic", "model": "claude-sonnet-4"}
			}
		},
		"roles": {
			"writer": {"llm_provider": "stub"}
		}
	}`)

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.Run.Workers != 1 {
		t.Fatalf("workers = %d, want default 1", cfg.Run.Workers)
	}
	if cfg.Run.MaxAttempts != 1 {
		t.Fatalf("max attempts = %d, want default 1", cfg.Run.MaxAttempts)
	}
	if cfg.Queue.Type != "inmemory" {
		t.Fatalf("queue type = %q, want default inmemory", cfg.Queue.Type)
	}
	if cfg.Queue.LeaseTimeout.Duration != 30*time.Second {
		t.Fatalf("lease timeout = %s, want default 30s", cfg.Queue.LeaseTimeout.Duration)
	}
}

func TestLoadFileRejectsUnknownExtension(t *testing.T) {
	path := writeTempConfig(t, "tessera.toml", "run = {}\n")

	_, err := LoadFile(path)
	if !errors.Is(err, ErrUnsupportedExtension) {
		t.Fatalf("LoadFile() error = %v, want ErrUnsupportedExtension", err)
	}
}

func TestValidateRejectsInvalidProviderReferences(t *testing.T) {
	path := writeTempConfig(t, "tessera.yaml", `
llm:
  default_provider: missing
  providers:
    claude:
      provider: anthropic
      model: claude-sonnet-4
roles:
  writer:
    llm_provider: codex
`)

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("LoadFile() error = nil, want validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "llm.default_provider") || !strings.Contains(msg, "roles.writer.llm_provider") {
		t.Fatalf("LoadFile() error = %v, want missing provider details", err)
	}
}

func writeTempConfig(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
