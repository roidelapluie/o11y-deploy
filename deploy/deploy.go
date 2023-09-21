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

package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/roidelapluie/o11y-deploy/ansible"
	"github.com/roidelapluie/o11y-deploy/config"
	ansiblemodel "github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/modules"
	"golang.org/x/exp/maps"
)

type Deployer struct {
	cfg          *config.Config
	homeDeps     string
	logger       log.Logger
	ansibleDebug int
}

// NewDeployer creates a new Deployer with the given configuration and homeDeps.
func NewDeployer(logger log.Logger, cfg *config.Config, homeDeps string, ansibleDebug int) (*Deployer, error) {
	d := &Deployer{
		cfg:          cfg,
		homeDeps:     homeDeps,
		logger:       logger,
		ansibleDebug: ansibleDebug,
	}
	return d, d.validateConfig()
}

// Run executes the deployment process and returns an error if anything goes wrong.
func (d *Deployer) Run() error {
	// Validate the configuration before proceeding with the deployment
	err := d.validateConfig()
	if err != nil {
		return err
	}

	level.Debug(d.logger).Log("msg", "Starting deployment...")

	c := context.Background()

	// Check if the directory already exists
	if _, err := os.Stat(d.cfg.Global.DataDir); os.IsNotExist(err) {
		// If the directory does not exist, create it
		if err := os.MkdirAll(d.cfg.Global.DataDir, 0755); err != nil {
			level.Error(d.logger).Log("msg", "Error creating data directory", "err", err, "path", d.cfg.Global.DataDir)
			return err
		}
	} else {
		level.Error(d.logger).Log("msg", "Data directory present", "path", d.cfg.Global.DataDir)
	}

	moduleTargets := make(map[string][]labels.Labels)
	prometheusTargets := make(map[string]map[string][]labels.Labels)
	ruleGroups := []rulefmt.RuleGroup{}
	dashboards := []map[string]interface{}{}
	dashboardFiles := make(map[string][]byte)
	lb := labels.NewBuilder(labels.EmptyLabels())
	for _, targetGroup := range d.cfg.TargetGroups {
		tgs := make([]labels.Labels, 0)
		targets, err := PopulateTargets(d.logger, targetGroup.Targets, time.Duration(d.cfg.Global.SDSyncTime))
		if err != nil {
			return err
		}
		for _, t := range targets {
			for _, tg := range t.Targets {
				lb.Reset(labels.EmptyLabels())

				for ln, lv := range tg {
					lb.Set(string(ln), string(lv))
				}
				for ln, lv := range t.Labels {
					if _, ok := tg[ln]; !ok {
						lb.Set(string(ln), string(lv))
					}
				}
				if relabeled, keep := relabel.Process(lb.Labels(labels.EmptyLabels()), targetGroup.Targets.RelabelConfigs...); keep {
					tgs = append(tgs, relabeled)
				}
			}
		}
		moduleTargets[targetGroup.Name] = tgs
		promTargets := make(map[string][]labels.Labels)

		for _, mod := range targetGroup.Modules.ModulesConfigs {
			if !mod.IsEnabled() {
				continue
			}
			m, err := mod.NewModule(modules.ModuleOptions{})
			if err != nil {
				return err
			}
			mtgs, err := m.GetTargets(tgs, targetGroup.Name)
			if err != nil {
				return err
			}
			promTargets[mod.Name()] = append(promTargets[mod.Name()], mtgs...)
			ruleGroups = append(ruleGroups, m.GetRules(targetGroup.Name))
			dashboards = append(dashboards, m.GetDashboards()...)
			maps.Copy(dashboardFiles, m.GetDashboardFiles())
		}
		prometheusTargets[targetGroup.Name] = promTargets
		c = ctx.SetPromRules(c, targetGroup.Name, ruleGroups)
	}

	c = ctx.SetPromTargets(c, prometheusTargets)
	c = ctx.SetDashboards(c, dashboards)
	c = ctx.SetDashboardFiles(c, dashboardFiles)

	for _, targetGroup := range d.cfg.TargetGroups {
		tgs, _ := moduleTargets[targetGroup.Name]

		inventory, err := generateInventory(tgs)
		if err != nil {
			return err
		}

		if targetGroup.Name == "prometheus" {
			servers := []string{}
			for host := range inventory.Groups["all"].Hosts {
				servers = append(servers, host)
			}
			c = ctx.SetPromServers(c, servers)
		}

		var pbs = make([]*ansiblemodel.Playbook, 0)
		for _, mod := range targetGroup.Modules.ModulesConfigs {
			if !mod.IsEnabled() {
				continue
			}
			m, err := mod.NewModule(modules.ModuleOptions{})
			if err != nil {
				return err
			}
			pb, err := m.Playbook(c)
			if err != nil {
				return err
			}
			pbs = append(pbs, pb)
		}
		ar, err := ansible.NewRunner(d.logger, d.cfg, d.ansibleDebug, filepath.Join(d.homeDeps, "bin", "ansible-playbook"), d.homeDeps, inventory)
		if err != nil {
			return err
		}
		err = ar.RunPlaybooks(c, pbs)
		if err != nil {
			return err
		}
	}

	level.Info(d.logger).Log("msg", "Deployment done")
	return nil
}

// validateConfig checks the Deployer's configuration for any issues.
func (d *Deployer) validateConfig() error {
	if d.cfg == nil {
		return errors.New("configuration cannot be nil")
	}

	if len(d.cfg.TargetGroups) == 0 {
		return errors.New("configuration must have at least one target group")
	}

	return nil
}

// PopulateTargets populates targets
func PopulateTargets(logger log.Logger, targets *config.Targets, syncTime time.Duration) ([]*targetgroup.Group, error) {
	targetGroupChan := make(chan []*targetgroup.Group)
	level.Info(logger).Log("msg", "Waiting", "time", syncTime)
	ctx, cancel := context.WithTimeout(context.Background(), syncTime)
	defer cancel()

	for _, cfg := range targets.ServiceDiscoveryConfigs {
		d, err := cfg.NewDiscoverer(discovery.DiscovererOptions{Logger: logger})
		if err != nil {
			return nil, fmt.Errorf("could not create new discoverer: %w", err)
		}
		go d.Run(ctx, targetGroupChan)
	}

	targetGroups := make([]*targetgroup.Group, 0)
	targetGroupsMap := make(map[string]*targetgroup.Group)
outerLoop:
	for {
		select {
		case targetGroups = <-targetGroupChan:
			for _, tg := range targetGroups {
				targetGroupsMap[tg.Source] = tg
			}
		case <-ctx.Done():
			break outerLoop
		}
	}

	var targetGroupsOutput []*targetgroup.Group
	for _, tgs := range targetGroupsMap {
		targetGroupsOutput = append(targetGroupsOutput, tgs)
	}

	return targetGroupsOutput, nil
}

func generateInventory(tgs []labels.Labels) (*ansiblemodel.Inventory, error) {
	group := ansiblemodel.Group{
		Hosts: map[string]ansiblemodel.Host{},
	}
	for _, tg := range tgs {
		if tg.IsEmpty() {
			continue
		}
		hostVars := map[string]interface{}{}
		name := tg.Get(model.AddressLabel)
		if name == "" {
			continue
		}

		group.Hosts[name] = ansiblemodel.Host{
			Variables: hostVars,
		}
	}

	inventory := ansiblemodel.Inventory{
		Groups: map[string]ansiblemodel.Group{
			"all": group,
		},
	}

	return &inventory, nil
}
