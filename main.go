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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"

	"github.com/roidelapluie/o11y-deploy/config"
	"github.com/roidelapluie/o11y-deploy/deploy"
)

const (
	expectedDepsVersion = "0.0.10"
)

var (
	depsHome     = kingpin.Flag("deps-home", "The home of O11y-deps").Default("/opt/o11y/deps").String()
	allowDepsDev = kingpin.Flag("allow-deps-dev", "Allow running with 'dev' version of dependencies").Bool()
	configFile   = kingpin.Flag("config-file", "The path to the configuration file").Default("o11y.yml").String()
	ara          = kingpin.Flag("ara", "Run the Ara webserver").Bool()
	ansibleDebug = kingpin.Flag("ansible.debug", "Run ansible in debug mode").Counter()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	absDepsHome, err := filepath.Abs(*depsHome)
	if err != nil {
		fmt.Println("Error converting the provided path to an absolute path:", err)
		os.Exit(1)
	}

	if absDepsHome == "/" {
		fmt.Println("Error: The provided path is forbidden")
		os.Exit(1)
	}

	depsHome = &absDepsHome

	if err := preflightCheck(*depsHome); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err := config.LoadFile(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if *ara {
		if !cfg.Global.EnableARA {
			fmt.Printf("Ara is disabled in the configuration")
			os.Exit(1)
		}

		cmd := exec.Command(filepath.Join(*depsHome, "bin", "ara-manage"), "migrate")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("ARA_DATABASE_NAME=%s", filepath.Join(cfg.Global.DataDir, "ansible.sqlite")))

		err = cmd.Run()
		if err != nil {
			logger.Log("msg", "Error running ara", "err", err)
			os.Exit(1)
		}

		cmd = exec.Command(filepath.Join(*depsHome, "bin", "ara-manage"), "runserver", cfg.Global.ARAListen)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("ARA_DATABASE_NAME=%s", filepath.Join(cfg.Global.DataDir, "ansible.sqlite")))

		err = cmd.Run()
		if err != nil {
			logger.Log("msg", "Error running ara", "err", err)
			os.Exit(1)
		}

		return
	}

	deployer, err := deploy.NewDeployer(logger, cfg, absDepsHome, *ansibleDebug)
	if err != nil {
		fmt.Printf("Error creating deployer: %v\n", err)
		os.Exit(1)
	}

	if err = deployer.Run(); err != nil {
		fmt.Printf("Error running deployer: %v\n", err)
		os.Exit(1)
	}
}

func preflightCheck(depsHome string) error {
	versionFilePath := filepath.Join(depsHome, "O11YDEPSVERSION")
	data, err := ioutil.ReadFile(versionFilePath)

	if err != nil {
		return fmt.Errorf("Error reading O11YDEPSVERSION file: %v", err)
	}

	existingVersion := strings.TrimSpace(string(data))

	if existingVersion != expectedDepsVersion {
		if existingVersion == "dev" && *allowDepsDev {
			fmt.Println("Running with development version of dependencies.")
		} else {
			return fmt.Errorf("Mismatched dependency version. Expected: %s, Found: %s", expectedDepsVersion, existingVersion)
		}
	}

	return nil
}
