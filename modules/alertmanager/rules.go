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

package alertmanager

import (
	"fmt"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// GetRules returns recording and alerting rules for this module.
func (m *Module) GetRules(tg string) rulefmt.RuleGroup {
	return rulefmt.RuleGroup{
		Name:  fmt.Sprintf("%v-alertmanager", tg),
		Rules: []rulefmt.RuleNode{},
	}
}
