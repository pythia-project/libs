// Pythia utility functions
// Author: Sébastien Combéfis <sebastien@combefis.be>
//
// Copyright (C) 2020, Computer Science and IT in Education ASBL
// Copyright (C) 2020, ECAM Brussels Engineering School
//
// This program is free software: you can redistribute it and/or modify
// under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 2 of the License, or
//  (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ExecutionResult contains the result of the execution of a process.
type ExecutionResult struct {
	ReturnCode int    `json:"returncode"`
	StdOut     string `json:"stdout"`
	StdErr     string `json:"stderr"`
}

const (
	WORKDIR = "/tmp/work"
)

// Setup working directory.
func SetupWorkDir() error {
	os.RemoveAll(WORKDIR)
	if err := os.MkdirAll(WORKDIR, 0777); err != nil {
		return err
	}
	return nil
}

// Read all data from the standard input.
func ReadStdIn() ([]byte, error) {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return bytes.TrimRight(input, "\u0000"), nil
}

func getExitStatus(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}

// Execute a command and retrieve execution results.
func Execute(command *string) ExecutionResult {
	var execResult ExecutionResult
	var stdout, stderr bytes.Buffer

	// Build the command to run
	tokens := strings.Split(*command, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command and retrieve execution results.
	if err := cmd.Run(); err != nil {
		execResult.ReturnCode = getExitStatus(err)
	}
	execResult.StdOut = string(stdout.Bytes())
	execResult.StdErr = string(stderr.Bytes())

	return execResult
}
