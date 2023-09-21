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

package ansible

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/roidelapluie/o11y-deploy/config"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/util/writer"

	"gopkg.in/yaml.v2"
)

//go:embed roles.tar.gz
var roles []byte

type AnsibleRunner struct {
	Logger      log.Logger
	AnsiblePath string
	DepsPath    string
	Inventory   *ansible.Inventory
	Config      *config.Config
	debug       int
	araPath     string
}

func NewRunner(logger log.Logger, cfg *config.Config, debug int, ansiblePath, depsPath string, inventory *ansible.Inventory) (*AnsibleRunner, error) {
	i := *inventory

	if _, ok := i.Groups["all"]; !ok {
		i.Groups["all"] = ansible.Group{}
	}
	allGr, ok := i.Groups["all"]
	if !ok {
		panic("group should exist at this point")
	}
	if allGr.Variables == nil {
		allGr.Variables = make(map[string]interface{})
	}

	if cfg.Global.AnsibleTOFU {
		level.Debug(logger).Log("msg", "Ansible TOFU enabled")
		allGr.Variables["ansible_ssh_extra_args"] = fmt.Sprintf("-o UserKnownHostsFile=%q -o StrictHostKeyChecking=no", filepath.Join(cfg.Global.DataDir, "known_hosts"))
	}
	allGr.Variables["ansible_user"] = cfg.Global.AnsibleUser
	allGr.Variables["ansible_ssh_private_key_file"] = cfg.Global.AnsibleSSHKeyPath
	if cfg.Global.AnsibleBecomePasswordFile != "" {
		d, err := os.ReadFile(cfg.Global.AnsibleBecomePasswordFile)
		if err != nil {
			return nil, err
		}
		allGr.Variables["ansible_become_pass"] = strings.TrimSpace(string(d))
	}

	i.Groups["all"] = allGr

	data, err := yaml.Marshal(&i)
	if err == nil {
		level.Debug(logger).Log("msg", "Ansible inventory", "inventory", string(data))
	}

	return &AnsibleRunner{
		Logger:      logger,
		AnsiblePath: ansiblePath,
		DepsPath:    depsPath,
		Inventory:   &i,
		Config:      cfg,
		debug:       debug,
	}, nil
}

func (ar *AnsibleRunner) FindARAPath() (string, error) {
	if ar.araPath != "" {
		return ar.araPath, nil
	}
	cmd := exec.Command(filepath.Join(ar.DepsPath, "bin", "python3"), "-m", "ara.setup.callback_plugins")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Command failed with error:", err)
		fmt.Println("Stderr:", stderr.String())
		return "", err
	}
	ar.araPath = strings.TrimSpace(stdout.String())
	if ar.araPath == "" {
		fmt.Println("Stderr:", stderr.String())
		return "", errors.New("No ARA Path!")
	}
	return ar.araPath, nil
}

func (ar *AnsibleRunner) RunPlaybooks(ctx context.Context, playbooks []*ansible.Playbook, extraArgs ...string) error {
	if len(playbooks) == 0 {
		return errors.New("No playbooks!")
	}
	rolesPath, err := extractGz(roles)
	if err != nil {
		return err
	}
	defer func() {
		if ar.debug > 0 {
			fmt.Printf("DEBUG: Roles kept in %q\n", rolesPath)
			return
		}
		os.RemoveAll(rolesPath)
	}()

	homeDir, err := os.MkdirTemp("", "home")
	if err != nil {
		return err
	}
	defer os.RemoveAll(homeDir)
	if err := os.MkdirAll(filepath.Join(homeDir, ".ara", "server", "www", "static"), 0777); err != nil {
		return err
	}

	cfgFile, err := ar.writeConfig(rolesPath)
	if err != nil {
		return err
	}
	defer func() {
		if ar.debug > 0 {
			fmt.Printf("DEBUG: config kept in %q\n", cfgFile)
			return
		}
		os.Remove(cfgFile)
	}()

	inventoryFile, err := write(ar.Inventory)
	if err != nil {
		return err
	}
	defer func() {
		if ar.debug > 0 {
			fmt.Printf("DEBUG: inventory kept in %q\n", inventoryFile)
			return
		}
		os.Remove(inventoryFile)
	}()

	playbookFile, err := write(playbooks)
	if err != nil {
		return err
	}
	defer func() {
		if ar.debug > 0 {
			fmt.Printf("DEBUG: playbook kept in %q\n", playbookFile)
			return
		}
		os.Remove(playbookFile)
	}()

	callbackFile, err := os.CreateTemp("", "ansible_log.json")
	if err != nil {
		return err
	}
	defer func() {
		if ar.debug > 0 {
			fmt.Printf("DEBUG: callback kept in %q\n", callbackFile.Name())
			return
		}
		os.Remove(callbackFile.Name())
	}()

	args := []string{"-i", inventoryFile, playbookFile}
	//if len(playbook.Targets) > 0 {
	//	args = append(args, "--limit", strings.Join(playbook.Targets, ","))
	//}
	args = append(args, extraArgs...)

	cmd := exec.Command(ar.AnsiblePath, args...)
	cmdWriter := writer.New(os.Stdout)
	errWriter := writer.NewBufferedWriter(cmdWriter)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if ar.debug == 0 {
		cmd.Stdout = cmdWriter
		cmd.Stderr = errWriter
	}
	cmd.Env = os.Environ()
	if ar.debug > 2 {
		cmd.Env = append(cmd.Env, "ANSIBLE_DEBUG=1")
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", homeDir))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_CONFIG=%s", cfgFile))
	cmd.Env = append(cmd.Env, "ANSIBLE_CALLBACKS_ENABLED=json_logger")
	cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_JSON_LOG_FILE=%s", callbackFile.Name()))

	callbackPath := filepath.Join(ar.DepsPath, "opt", "o11y", "ansible_callback")
	if ar.Config.Global.EnableARA {
		araPath, err := ar.FindARAPath()
		if err != nil {
			ar.Logger.Log("msg", "Error getting ARA Path", "err", err)
			return err
		}
		if ar.debug > 0 {
			fmt.Printf("ARA Enabled at %q\n", araPath)
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_CALLBACK_PLUGINS=%s:%s", araPath, callbackPath))
		cmd.Env = append(cmd.Env, fmt.Sprintf("ARA_DATABASE_NAME=%s", filepath.Join(ar.Config.Global.DataDir, "ansible.sqlite")))
	} else {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_CALLBACK_PLUGINS=%s", callbackPath))
	}
	if ar.debug > 1 {
		for _, s := range cmd.Env {
			fmt.Println(s)
		}
	}

	err = cmd.Run()

	if ar.debug == 0 {
		errWriter.WriteAll(os.Stderr)
	}

	readAndPrintJSONReport(callbackFile.Name())

	if err != nil {
		ar.Logger.Log("msg", "Error running playbook", "err", err)
		return err
	}

	ar.Logger.Log("msg", "Playbook execution completed")

	return nil
}

func write(i interface{}) (string, error) {
	tempFile, err := os.CreateTemp("", "o11y_")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	data, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	if _, err := tempFile.Write(data); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func (ar *AnsibleRunner) writeConfig(rolesPath string) (string, error) {
	cfgFile, err := os.CreateTemp("", "ansible*.cfg")
	if err != nil {
		return "", err
	}
	// Write the ansible.cfg content to the
	// temporary file
	cfgContent := fmt.Sprintf(`
[defaults]
roles_path = %s
`, rolesPath)

	if _, err := cfgFile.Write([]byte(cfgContent)); err != nil {
		return "", err
	}

	level.Debug(ar.Logger).Log("msg", "Ansible config", "content", cfgContent)

	return cfgFile.Name(), nil
}
