/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package setup

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"omg-cli/config"
	"omg-cli/ssh"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Jumpbox allows SSHing to the real Jumpbox.
type Jumpbox struct {
	logger  *log.Logger
	output  io.Writer
	session *ssh.Connection
	envDir  string
}

const packageName = "omg-cli"

// NewJumpbox creates a new Jumpbox.
func NewJumpbox(cmdLogger *log.Logger, output io.Writer, ip, username, sshKeyPath, envDir string, quiet bool) (*Jumpbox, error) {
	var logger *log.Logger
	if !quiet {
		// Duplicate the logger so we can modify the prefix
		logger = log.New(output, fmt.Sprintf("%s[jumpbox] ", cmdLogger.Prefix()), 0)
	} else {
		logger = log.New(ioutil.Discard, "", 0)
	}
	key, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		return nil, err
	}

	jb := &Jumpbox{logger: logger, output: output, envDir: envDir}
	jb.session, err = ssh.NewConnection(logger, output, ip, ssh.Port, username, key)
	if err != nil {
		return nil, err
	}

	return jb, nil
}

// PoolTillStarted waits until the jumpbox is listening for SSH connections.
func (jb *Jumpbox) PoolTillStarted() error {
	timer := time.After(time.Duration(0 * time.Second))
	timeout := time.After(time.Duration(120 * time.Second))
	for {
		select {
		case <-timeout:
			return errors.New("timeout waiting for Jumpbox to start")
		case <-timer:
			if err := jb.session.EnsureConnected(); err == nil {
				return nil
			}
			jb.logger.Print("waiting for Jumpbox to start")
			timer = time.After(time.Duration(5 * time.Second))
		}
	}
}

// UploadDependencies pushes the OMG binary, environment config to the jumpbox.
func (jb *Jumpbox) UploadDependencies() error {
	if err := jb.session.EnsureConnected(); err != nil {
		return err
	}

	rebuilt, err := ioutil.TempFile("", "tile")
	if err != nil {
		return err
	}
	defer os.Remove(rebuilt.Name())
	build := exec.Command("go", "build", "-o", rebuilt.Name())
	build.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	build.Dir = filepath.Join(".", "src", "omg-cli")

	build.Stderr = os.Stderr
	build.Stdout = os.Stdout
	if err := build.Run(); err != nil {
		return fmt.Errorf("rebuilding go: %v - environment: %v", err, build.Env)
	}

	type plan struct {
		local, dest string
	}
	files := []plan{
		// Upload the omg-cli binary to be run remotely.
		{rebuilt.Name(), packageName},
		// Upload the jumpbox public SSH key for director-ssh cmd.
		{filepath.Join(jb.envDir, "keys", "jumpbox_ssh.pub"), "keys/jumpbox_ssh.pub"},
	}

	for _, f := range config.Files {
		files = append(files, plan{filepath.Join(jb.envDir, f), f})
	}

	if err := jb.session.Mkdir("keys"); err != nil {
		return err
	}
	for _, f := range files {
		if err := jb.session.UploadFile(f.local, f.dest); err != nil {
			return fmt.Errorf("uploading file %s: %v", f.local, err)
		}
	}

	return nil
}

// RunOmg runs an omg-cli command on the jumpbox.
func (jb *Jumpbox) RunOmg(args string) error {
	if err := jb.session.EnsureConnected(); err != nil {
		return err
	}

	return jb.session.RunCommand(fmt.Sprintf("~/%s %s --env-dir=$PWD", packageName, args))
}
