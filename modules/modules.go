// Copyright 2020 The Prometheus Authors
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

package modules

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"reflect"

	"github.com/go-kit/log"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
)

// Module is the interface for modules.
type Module interface {
	Playbook(context.Context) (*ansible.Playbook, error)
	HostVars(target labels.Labels, group string) (map[string]interface{}, error)
	GetTargets([]labels.Labels, string) ([]labels.Labels, error)
	GetRules(string) rulefmt.RuleGroup
	GetDashboards() []map[string]interface{}
	GetDashboardFiles() map[string][]byte
}

// ModuleOptions provides options for a Module.
type ModuleOptions struct {
	Logger log.Logger
}

// A Config provides the configuration and constructor for a Module.
type Config interface {
	// Name returns the name of the discovery mechanism.
	Name() string

	// IsEnabled returns a boolean indicating if the Module is enabled.
	IsEnabled() bool

	// NewModule returns a Discoverer for the Config
	// with the given DiscovererOptions.
	NewModule(ModuleOptions) (Module, error)
}

// Configs is a slice of Config values that uses custom YAML marshaling and unmarshaling
// to represent itself as a mapping of the Config values grouped by their types.
type Configs []Config

// SetDirectory joins any relative file paths with dir.
func (c *Configs) SetDirectory(dir string) {
	for _, c := range *c {
		if v, ok := c.(config.DirectorySetter); ok {
			v.SetDirectory(dir)
		}
	}
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Configs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	cfgTyp := getConfigType(configsType)
	cfgPtr := reflect.New(cfgTyp)
	cfgVal := cfgPtr.Elem()

	if err := unmarshal(cfgPtr.Interface()); err != nil {
		return replaceYAMLTypeError(err, cfgTyp, configsType)
	}

	var err error
	*c, err = readConfigs(cfgVal, 0)
	return err
}

// MarshalYAML implements yaml.Marshaler.
func (c Configs) MarshalYAML() (interface{}, error) {
	cfgTyp := getConfigType(configsType)
	cfgPtr := reflect.New(cfgTyp)
	cfgVal := cfgPtr.Elem()

	if err := writeConfigs(cfgVal, c); err != nil {
		return nil, err
	}

	return cfgPtr.Interface(), nil
}

// GetTargets transforms targets into prometheus targets
func GetTargets(targets []labels.Labels, port, group string) ([]labels.Labels, error) {
	promTargets := make([]labels.Labels, 0)
	for _, target := range targets {
		t := target.Copy()
		addr := t.Get(model.AddressLabel)
		if addr == "" {
			return nil, fmt.Errorf("__address__ label not found in label set")
		}
		host, _, err := net.SplitHostPort(string(addr))
		if err != nil {
			host = string(addr)
		}
		lbs := labels.NewBuilder(t)
		lbs.Set(model.AddressLabel, net.JoinHostPort(host, port))
		lbs.Set("group_name", group)
		promTargets = append(promTargets, lbs.Labels(labels.EmptyLabels()))
	}
	return promTargets, nil
}

// GetReverseProxy transforms targets into reverse proxy entries
func GetReverseProxy(targets []labels.Labels, port, name, prefix, group string) ([]ReverseProxyEntry, error) {
	entries := make([]ReverseProxyEntry, len(targets))
	for i, target := range targets {
		t := target.Copy()
		addr := t.Get(model.AddressLabel)
		if addr == "" {
			return nil, fmt.Errorf("__address__ label not found in label set")
		}
		host, _, err := net.SplitHostPort(string(addr))
		if err != nil {
			host = string(addr)
		}
		ename := fmt.Sprintf("%s on %s", name, host)
		entries[i] = ReverseProxyEntry{
			Name:   name,
			URL:    fmt.Sprintf("http://%s", net.JoinHostPort(host, port)),
			Prefix: prefix + "/" + HashString(ename),
		}

	}
	return entries, nil
}

func GetReverseProxyAddress(target labels.Labels, name, prefix, group string) (string, error) {
	t := target.Copy()
	addr := t.Get(model.AddressLabel)
	if addr == "" {
		return "", fmt.Errorf("__address__ label not found in label set")
	}
	host, _, err := net.SplitHostPort(string(addr))
	if err != nil {
		host = string(addr)
	}
	ename := fmt.Sprintf("%s on %s", name, host)
	return "{{o11y_portal_address|default(\"\")}}" + prefix + "/" + HashString(ename), nil
}

func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))[:20]
}
