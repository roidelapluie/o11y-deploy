// Copyright 2023 The O11y Authors and The Prometheus Authors
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

//go:build !builtinassets
// +build !builtinassets

package ui

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/httpfs/union"
)

var Assets = func() http.FileSystem {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var assetsPrefix string
	switch filepath.Base(wd) {
	case "o11y-deploy":
		// When running o11y-deploy (without built-in assets) from the repo root.
		assetsPrefix = "./frontend/ui/build"
	case "frontend":
		// When running frontend tests.
		assetsPrefix = "./ui/build"
	case "ui":
		// When generating statically compiled assets.
		assetsPrefix = "./build"
	}

	static := filter.Keep(
		http.Dir(path.Join(assetsPrefix, "static")),
		func(path string, fi os.FileInfo) bool {
			return fi.IsDir()
		},
	)

	return union.New(map[string]http.FileSystem{
		"/static": static,
	})
}()
