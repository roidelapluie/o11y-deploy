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
)

type key int

const (
	promTargets key = iota
	promRules   key = iota
)

func GetPromTargets(ctx context.Context) map[string][]labels.Labels {
	targets, ok := ctx.Value(promTargets).(map[string][]labels.Labels)
	if !ok {
		return nil
	}
	return targets
}

func SetPromTargets(ctx context.Context, targets map[string][]labels.Labels) context.Context {
	return context.WithValue(ctx, promTargets, targets)
}

// GetPromRules gets Prometheus rule groups from the context
func GetPromRules(ctx context.Context) []rulefmt.RuleGroup {
	rules, ok := ctx.Value(promRules).([]rulefmt.RuleGroup)
	if !ok {
		return nil
	}
	return rules
}

// SetPromRules adds Prometheus rule groups to the context
func SetPromRules(ctx context.Context, rules []rulefmt.RuleGroup) context.Context {
	return context.WithValue(ctx, promRules, rules)
}
