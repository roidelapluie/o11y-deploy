// Copyright 2023 The O11y Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/modules"
	"gopkg.in/yaml.v3"
)

var DefaultConfig = ModuleConfig{
	Enabled:           false,
	PrometheusVersion: "2.43.0",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	Enabled           bool   `yaml:"enabled"`
	PrometheusVersion string `yaml:"prometheus_version"`
}

func (m *ModuleConfig) Name() string {
	return "prometheus"
}

type ScrapeConfig struct {
	JobName        string         `yaml:"job_name"`
	ScrapeInterval time.Duration  `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration  `yaml:"scrape_timeout"`
	MetricsPath    string         `yaml:"metrics_path"`
	Scheme         string         `yaml:"scheme"`
	StaticConfigs  []StaticConfig `yaml:"static_configs"`
}

type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (m *ModuleConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*m = DefaultConfig
	// We want to set m to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.
	type plain ModuleConfig
	if err := unmarshal((*plain)(m)); err != nil {
		return err
	}
	return nil
}

func (m *ModuleConfig) NewModule(modules.ModuleOptions) (modules.Module, error) {
	return &Module{
		cfg: m,
	}, nil
}

// IsEnabled returns a boolean indicating if the Module is enabled.
func (m *ModuleConfig) IsEnabled() bool {
	return m.Enabled
}

type Module struct {
	cfg *ModuleConfig
}

func (m *Module) Playbook(c context.Context) (*ansible.Playbook, error) {
	staticConfig := labelsToStaticConfigs(ctx.GetPromTargets(c))

	scrapeConfigs := []*ScrapeConfig{
		{
			JobName:        "example",
			ScrapeInterval: 15 * time.Second,
			ScrapeTimeout:  10 * time.Second,
			MetricsPath:    "/metrics",
			Scheme:         "http",
			StaticConfigs:  staticConfig,
		},
	}

	rulesFile, err := os.CreateTemp("", "prometheus_*.rules")
	if err != nil {
		return nil, fmt.Errorf("could not create rules file: %v", err)
	}

	ruleGroups := rulefmt.RuleGroups{
		Groups: ctx.GetPromRules(c),
	}

	rulesYaml, err := yaml.Marshal(ruleGroups)
	if err != nil {
		return nil, fmt.Errorf("could not marshal rules file: %v", err)
	}

	if _, err := rulesFile.Write([]byte(rulesYaml)); err != nil {
		return nil, fmt.Errorf("could not write rules file: %v", err)
	}

	rulesLink := filepath.Join(os.TempDir(), "prometheus.rules")
	if _, err := os.Stat(rulesLink); err == nil {
		// Link exists; delete it so we can relink to the new version
		os.Remove(rulesLink)
	}

	err = os.Link(rulesFile.Name(), rulesLink)
	if err != nil {
		return nil, fmt.Errorf("could not link rules file: %v", err)
	}

	return &ansible.Playbook{
		Name: "Linux",
		Vars: map[string]interface{}{
			"prometheus_version":           m.cfg.PrometheusVersion,
			"prometheus_scrape_configs":    scrapeConfigs,
			"prometheus_alert_rules":       []string{},
			"prometheus_alert_rules_files": []string{rulesLink},
		},
		Hosts:  "all",
		Become: true,
		Roles: []ansible.Role{
			ansible.Role{
				Name: "prometheus",
			},
		},
	}, nil
}

func (m *Module) HostVars() (map[string]string, error) {
	return nil, nil
}

func (m *Module) GetTargets(targets []labels.Labels) ([]labels.Labels, error) {
	return nil, nil
}

func labelsToStaticConfigs(labelSetList map[string][]labels.Labels) []StaticConfig {
	staticConfigs := make([]StaticConfig, 0, len(labelSetList))

	for _, l := range labelSetList {
		for _, labelSet := range l {
			staticConfig := StaticConfig{
				Targets: make([]string, 0, 1),
				Labels:  make(map[string]string, len(labelSet)),
			}

			instance := ""
			for _, label := range labelSet {
				if label.Name == model.AddressLabel {
					instance = label.Value
				} else {
					staticConfig.Labels[label.Name] = label.Value
				}
			}
			staticConfig.Targets = append(staticConfig.Targets, instance)
			staticConfigs = append(staticConfigs, staticConfig)
		}
	}

	return staticConfigs
}
