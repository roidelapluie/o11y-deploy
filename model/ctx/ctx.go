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

package ctx

import (
	"context"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/roidelapluie/o11y-deploy/modules"
)

type key int

const (
	targetGroup           key = iota
	promTargets           key = iota
	promRules             key = iota
	promServers           key = iota
	grafanaDashboards     key = iota
	grafanaDashboardFiles key = iota
	dataDir               key = iota
	reverseProxyEntries   key = iota
)

func GetPromTargets(ctx context.Context) map[string]map[string][]labels.Labels {
	targets, ok := ctx.Value(promTargets).(map[string]map[string][]labels.Labels)
	if !ok {
		return nil
	}
	return targets
}

func SetPromTargets(ctx context.Context, targets map[string]map[string][]labels.Labels) context.Context {
	return context.WithValue(ctx, promTargets, targets)
}

// GetPromRules gets Prometheus rule groups from the context
func GetPromRules(ctx context.Context) map[string][]rulefmt.RuleGroup {
	rules, ok := ctx.Value(promRules).(map[string][]rulefmt.RuleGroup)
	if !ok {
		return nil
	}
	return rules
}

// SetPromRules adds Prometheus rule groups to the context
func SetPromRules(ctx context.Context, tg string, rules []rulefmt.RuleGroup) context.Context {
	rulesMap, ok := ctx.Value(promRules).(map[string][]rulefmt.RuleGroup)
	if !ok {
		rulesMap = make(map[string][]rulefmt.RuleGroup)
	}
	rulesMap[tg] = rules
	return context.WithValue(ctx, promRules, rulesMap)
}

// GetPromServers gets Prometheus server IP's from the context
func GetPromServers(ctx context.Context) []string {
	servers, ok := ctx.Value(promServers).([]string)
	if !ok {
		return nil
	}
	return servers
}

// SetPromServers adds Prometheus server IP's to the context
func SetPromServers(ctx context.Context, servers []string) context.Context {
	return context.WithValue(ctx, promServers, servers)
}

// GetDashboards gets Grafana dashboards from the context
func GetDashboards(ctx context.Context) []map[string]interface{} {
	dashboards, ok := ctx.Value(grafanaDashboards).([]map[string]interface{})
	if !ok {
		return nil
	}
	return dashboards
}

// SetDashboards adds dashboards to the context
func SetDashboards(ctx context.Context, dashboards []map[string]interface{}) context.Context {
	return context.WithValue(ctx, grafanaDashboards, dashboards)
}

func SetDashboardFiles(ctx context.Context, dashboards map[string][]byte) context.Context {
	return context.WithValue(ctx, grafanaDashboardFiles, dashboards)
}

// GetDashboardFiles gets Grafana dashboard files from the context
func GetDashboardFiles(ctx context.Context) map[string][]byte {
	dashboards, ok := ctx.Value(grafanaDashboardFiles).(map[string][]byte)
	if !ok {
		return nil
	}
	return dashboards
}

func SetDatadir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, dataDir, dir)
}

// GetDatadir gets data dir from the context
func GetDatadir(ctx context.Context) string {
	dir, ok := ctx.Value(dataDir).(string)
	if !ok {
		return ""
	}
	return dir
}

func SetReverseProxyEntries(ctx context.Context, entries []modules.ReverseProxyEntry) context.Context {
	return context.WithValue(ctx, reverseProxyEntries, entries)
}

func GetReverseProxyEntries(ctx context.Context) []modules.ReverseProxyEntry {
	entries, ok := ctx.Value(reverseProxyEntries).([]modules.ReverseProxyEntry)
	if !ok {
		return nil
	}
	return entries
}
