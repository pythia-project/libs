// Pythia utilities for tasks execution
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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ExecutionResult contains the result of the execution of a program.
type ExecutionResult struct {
	ReturnCode int    `json:"returncode"`
	StdOut     string `json:"stdout"`
	StdErr     string `json:"stderr"`
}

const (
	workDir = "/tmp/work"
)

func main() {
	var tokens []string
	var execResult ExecutionResult

	// Parse arguments
	fileName := flag.String("filename", "", "Program source code file name.")
	compileCmd := flag.String("compile", "", "Command to compile the program.")
	executeCmd := flag.String("execute", "", "Command to execute the program.")
	flag.Parse()

	// Setup working directory
	os.RemoveAll(workDir)
	if err := os.MkdirAll(workDir, 0777); err != nil {
		log.Fatalf("Error while creating working directory: %s.", err)
	}

	// Read input data
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Error while reading stdin: %s.", err)
	}
	input = bytes.TrimRight(input, "\u0000")

	// Create source code file
	srcFile := fmt.Sprintf("%s/%s", workDir, *fileName)
	if err := ioutil.WriteFile(srcFile, input, 0774); err != nil {
		log.Fatalf("Error while creating source code file: %s.", err)
	}

	// Compile program
	if *compileCmd != "" {
		var stdout, stderr bytes.Buffer
		tokens = strings.Split(*compileCmd, " ")
		cmd := exec.Command(tokens[0], tokens[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					execResult.ReturnCode = status.ExitStatus()
				}
			}

			if stderr := stderr.Bytes(); len(stderr) > 0 {
				execResult.StdErr = string(stderr)
				result, err := json.Marshal(execResult)
				if err != nil {
					log.Fatalf("Error while generating JSON output: %s.", err)
				}
				fmt.Println(string(result))
				os.Exit(0)
			}
			log.Fatalf("Error while executing the compilation command: %s.", err)
		}
	}

	// Execute program
	if *executeCmd != "" {
		var stdout, stderr bytes.Buffer
		tokens = strings.Split(*executeCmd, " ")
		cmd := exec.Command(tokens[0], tokens[1:]...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					execResult.ReturnCode = status.ExitStatus()
				}
			}

			if stderr := stderr.Bytes(); len(stderr) > 0 {
				execResult.StdErr = string(stderr)
				result, err := json.Marshal(execResult)
				if err != nil {
					log.Fatalf("Error while generating JSON output: %s.", err)
				}
				fmt.Println(string(result))
				os.Exit(0)
			}
			if stdout := stdout.Bytes(); len(stdout) > 0 {
				execResult.StdOut = string(stdout)
				result, err := json.Marshal(execResult)
				if err != nil {
					log.Fatalf("Error while generating JSON output: %s.", err)
				}
				fmt.Println(string(result))
				os.Exit(0)
			}
			log.Fatalf("Error while executing the execution command: %s.", err)
		}

		execResult.StdOut = string(stdout.Bytes())
		result, err := json.Marshal(execResult)
		if err != nil {
			log.Fatalf("Error while generating JSON output: %s.", err)
		}
		fmt.Println(string(result))
	}

	os.Exit(0)
}
