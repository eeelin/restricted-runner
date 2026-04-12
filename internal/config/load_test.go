package config

import "testing"

func TestLoadYAMLAcceptsValidConfig(t *testing.T) {
	data := []byte(`
version: v1
root_path: /opt/restricted-runner/root
scripts:
  - path: homecloud/site/apply
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
    allow_argv: true
    allow_stdin: false
    allowed_env:
      - TARGET
    required_env:
      - TARGET
`)

	cfg, err := LoadYAML(data)
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
	if cfg.RootPath != "/opt/restricted-runner/root" {
		t.Fatalf("unexpected root path: %s", cfg.RootPath)
	}
}

func TestLoadYAMLRejectsInvalidConfig(t *testing.T) {
	data := []byte(`
version: v1
root_path: relative/root
scripts:
  - path: homecloud/site/apply
`)

	_, err := LoadYAML(data)
	if err == nil || err.Error() != "root_path must be absolute" {
		t.Fatalf("expected root_path error, got %v", err)
	}
}
