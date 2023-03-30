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

package main

import (
	_ "github.com/prometheus/prometheus/discovery/digitalocean"
	_ "github.com/prometheus/prometheus/discovery/file"
	_ "github.com/prometheus/prometheus/discovery/http"

	_ "github.com/roidelapluie/o11y-deploy/modules/linux"
	_ "github.com/roidelapluie/o11y-deploy/modules/prometheus"
)
