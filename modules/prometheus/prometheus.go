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
	"sort"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/model/promserver"
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
	scrapeConfigs := []ScrapeConfig{}
	staticConfigMap := labelsToStaticConfigs(ctx.GetPromTargets(c))

	for job, staticConfig := range staticConfigMap {
		scrapeConfigs = append(scrapeConfigs, ScrapeConfig{
			JobName:        job,
			ScrapeInterval: 15 * time.Second,
			ScrapeTimeout:  10 * time.Second,
			MetricsPath:    "/metrics",
			Scheme:         "http",
			StaticConfigs:  staticConfig,
		})
	}

	sort.Slice(scrapeConfigs, func(i, j int) bool {
		return scrapeConfigs[i].JobName < scrapeConfigs[j].JobName
	})

	rulesFiles := []string{}

	rulesMap := ctx.GetPromRules(c)
	rulesDir, err := os.MkdirTemp("", "prometheus_rules_*")
	if err != nil {
		return nil, fmt.Errorf("could not create rules directory: %v", err)
	}

	for tg, rules := range rulesMap {
		rulesTgDir := filepath.Join(rulesDir, tg)
		if err = os.Mkdir(rulesTgDir, 0750); err != nil {
			return nil, fmt.Errorf("could not create rules subdirectory: %v", err)
		}

		rulesFile, err := os.Create(filepath.Join(rulesTgDir, "prometheus.rules"))
		if err != nil {
			return nil, fmt.Errorf("could not create rules file: %v", err)
		}

		rulesLink := filepath.Join(os.TempDir(), fmt.Sprintf("%v.rules", tg))
		if _, err := os.Stat(rulesLink); err == nil {
			// Link exists; delete it so we can relink to the new version
			os.Remove(rulesLink)
		}

		err = os.Symlink(rulesFile.Name(), rulesLink)
		if err != nil {
			return nil, fmt.Errorf("could not link rules directory: %v", err)
		}

		rulesFiles = append(rulesFiles, rulesLink)

		ruleGroups := rulefmt.RuleGroups{
			Groups: rules,
		}

		rulesYaml, err := yaml.Marshal(ruleGroups)
		if err != nil {
			return nil, fmt.Errorf("could not marshal rules file: %v", err)
		}

		if _, err := rulesFile.Write([]byte(rulesYaml)); err != nil {
			return nil, fmt.Errorf("could not write rules file: %v", err)
		}

	}

	return &ansible.Playbook{
		Name: "Linux",
		Vars: map[string]interface{}{
			"prometheus_version":              m.cfg.PrometheusVersion,
			"prometheus_scrape_configs":       scrapeConfigs,
			"prometheus_alert_rules":          []string{},
			"prometheus_alert_rules_files":    rulesFiles,
			"prometheus_static_targets_files": []string{},
			"prometheus_web_external_url":     "{{o11y_prometheus_external_address}}",
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

func (m *Module) HostVars(target labels.Labels, group string) (map[string]interface{}, error) {
	addr, err := modules.GetReverseProxyAddress(target, m.cfg.Name(), "/prometheus", group)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"o11y_prometheus_external_address": addr,
	}, nil
}

func (m *Module) GetTargets(targets []labels.Labels, group string) ([]labels.Labels, error) {
	return modules.GetTargets(targets, "9090", group)
}

func (m *Module) ReverseProxy(targets []labels.Labels, group string) ([]modules.ReverseProxyEntry, error) {
	return modules.GetReverseProxy(targets, "9090", m.cfg.Name(), "/prometheus", group)
}

func (m *Module) GetPrometheusServers(targets []labels.Labels, group string) ([]promserver.PrometheusServer, error) {
	rp, err := modules.GetReverseProxy(targets, "9090", m.cfg.Name(), "/prometheus", group)
	if err != nil {
		return nil, err
	}
	promservers := []promserver.PrometheusServer{}
	for _, r := range rp {
		promservers = append(promservers, promserver.PrometheusServer{
			Name: r.Name,
			URL:  r.URL + r.Prefix,
		})
	}
	return promservers, nil
}

func labelsToStaticConfigs(labelSetList map[string]map[string][]labels.Labels) map[string][]StaticConfig {
	staticConfigsMap := make(map[string][]StaticConfig)

	for _, l := range labelSetList {
		for job, k := range l {
			if _, ok := staticConfigsMap[job]; !ok {
				staticConfigsMap[job] = make([]StaticConfig, 0, len(l))
			}
			for _, labelSet := range k {
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
				staticConfigsMap[job] = append(staticConfigsMap[job], staticConfig)
			}
		}
	}
	return staticConfigsMap
}
