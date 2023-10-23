package config

import (
	"testing"

	"gopkg.in/yaml.v2"

	_ "github.com/roidelapluie/o11y-deploy/modules/grafana"
	_ "github.com/roidelapluie/o11y-deploy/modules/linux"
	_ "github.com/roidelapluie/o11y-deploy/modules/prometheus"
)

func TestUnmarshalYAML(t *testing.T) {
	yamlString := `
global:
  ansible_ssh_key_path: /path/to/ansible_private_key.pem
  ansible_become_password_file: /path/to/ansible_password.txt
target_groups:
  - name: servers
    modules:
      linux_module:
      xlinux_module:
    targets:
      static_confis:
      - targets:
          - 'localhost:22'
        labels:
          group: 'o11y'
`
	var c Config
	err := yaml.Unmarshal([]byte(yamlString), &c)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if n := len(c.TargetGroups[0].Modules.ModulesConfigs); n != 1 {
		t.Fatalf("Expected 1 module, got %v", n)
	}
}
