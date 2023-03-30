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

package linux

import (
	"context"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/modules"
)

var DefaultConfig = ModuleConfig{
	PrometheusVersion: "2.43.0",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
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

	return &ansible.Playbook{
		Name: "Linux",
		Vars: map[string]interface{}{
			"prometheus_version":        m.cfg.PrometheusVersion,
			"prometheus_scrape_configs": scrapeConfigs,
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
