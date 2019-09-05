// Pythia utilities for input-output tasks
// Author: Sébastien Combéfis <sebastien@combefis.be>
//
// Copyright (C) 2019, Computer Science and IT in Education ASBL
// Copyright (C) 2019, ECAM Brussels Engineering School
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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TaskInput contains the inputs of the learner for the specified task id.
type TaskInput struct {
	Tid    string            `json:"tid"`
	Fields map[string]string `json:"fields"`
}

// TestConfig contains the configuration of the tests for a task.
type TestConfig struct {
	Predefined []struct {
		Input   string `json:"input"`
		Output  string `json:"output"`
		Message string `json:"message,omitempty"`
	} `json:"predefined"`
}

// TestOutput contains the output of the execution of the task.
type TestOutput struct {
	Results []Result `json:"results"`
}

// Result contains the result of one test.
type Result struct {
	Status string `json:"status"`
	Output string `json:"output"`
}

// Example contains a counterexample as a witness for a failed test.
type Example struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

// Stats contains statistical information about the tests execution.
type Stats struct {
	Succeeded int `json:"succeeded"`
	Total     int `json:"total"`
}

// Feedback contains feedback information about the tests execution.
type Feedback struct {
	Message string   `json:"message,omitempty"`
	Example *Example `json:"example,omitempty"`
	Stats   *Stats   `json:"stats,omitempty"`
	Score   float32  `json:"score"`
}

// Grading contains the result of the grading of the specified task id.
type Grading struct {
	Tid      string    `json:"tid"`
	Status   string    `json:"status"`
	Feedback *Feedback `json:"feedback,omitempty"`
}

const (
	skeletonDir = "/task/skeleton"

	workDir    = "/tmp/work"
	studentDir = workDir + "/student"
)

var fcts = map[string]func() error{
	"preprocess": preprocess,
	"execute":    execute,
	"feedback":   feedback,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Subcommand is required (preprocess, execute or feedback).")
	}

	// Find the function to execute for given subcommand.
	handler, ok := fcts[os.Args[1]]
	if !ok {
		log.Fatalf("Unknown subcommand: %s.", os.Args[1])
	}

	// Execute the function associated to the subcommand.
	if err := handler(); err != nil {
		log.Fatalf("Error while executing %s: %s.", os.Args[1], err)
	}
	os.Exit(0)
}

////////////////////////////////////////////////////////////////////////////////
// Preprocess

func preprocess() error {
	// Setup working directory.
	os.RemoveAll(workDir)
	if err := createDir(0755, workDir); err != nil {
		return err
	}
	if err := createDir(0777, studentDir, workDir+"/output"); err != nil {
		return err
	}

	// Read and parse input data.
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	input = strings.TrimRight(input, "\u0000")

	var data TaskInput
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return err
	}

	// Fill skeleton files with learner's inputs.
	if err := fillSkeletonFiles(skeletonDir, studentDir, data.Fields); err != nil {
		return err
	}

	// Save task id to file.
	if err := saveTaskId(data.Tid); err != nil {
		return err
	}

	return nil
}

func createDir(perm os.FileMode, paths ...string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, perm); err != nil {
				return err
			}
			if err := os.Chmod(path, perm); err != nil {
				return err
			}
		}
	}
	return nil
}

func fillSkeletonFiles(src string, dst string, fields map[string]string) error {
	// Check each file of the specified source directory.
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsDir() {
			// Get the destination file path.
			dstDir, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			dstFile := fmt.Sprintf("%s/%s", dst, dstDir)

			// Create destination directories.
			if err := createDir(0755, filepath.Dir(dstFile)); err != nil {
				return err
			}

			// Read the source file.
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			fileContent := string(content)

			// Find placeholders and replace them with corresponding input.
			for key, value := range fields {
				regex, _ := regexp.Compile("@([^@]*)@" + key + "@([^@]*)@")
				matches := regex.FindAllStringSubmatch(fileContent, -1)
				for _, match := range matches {
					lines := strings.Split(value, "\n")
					var rep strings.Builder
					for _, line := range lines {
						rep.WriteString(match[1] + line + match[2] + "\n")
					}
					fileContent = strings.ReplaceAll(fileContent, "@"+match[1]+"@"+key+"@"+match[2]+"@", rep.String())
				}
			}

			// Write the destination file.
			if err := ioutil.WriteFile(dstFile, []byte(fileContent), 0774); err != nil {
				return err
			}

			// Set file permission.
			if err := os.Chmod(dstFile, 0644); err != nil {
				return err
			}
		}
		return nil
	})
}

func saveTaskId(tid string) error {
	return ioutil.WriteFile(workDir+"/tid", []byte(tid), 0444)
}

////////////////////////////////////////////////////////////////////////////////
// Execute

func execute() error {
	if len(os.Args) < 3 {
		return errors.New("Command to execute is missing.")
	}

	// Read and parse test configuration.
	var config TestConfig
	if err := readTestConfig("/task/config/test.json", &config); err != nil {
		return err
	}

	// Execute the code from the learner.
	var output TestOutput
	output.Results = make([]Result, len(config.Predefined))
	for i, test := range config.Predefined {
		stdout, err := executeCommand(test.Input, os.Args[2], os.Args[3:]...)
		if err != nil {
			return err
		}
		tokens := strings.SplitN(stdout, "\n", 2)
		output.Results[i].Status = tokens[0]
		output.Results[i].Output = tokens[1]
	}

	// Write the produced output.
	resFile := workDir + "/output/res.json"
	file, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(resFile, file, 0644); err != nil {
		return err
	}

	// Set file permission.
	if err := os.Chmod(resFile, 0644); err != nil {
		return err
	}

	return nil
}

func readTestConfig(path string, config *TestConfig) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &config)
}

func executeCommand(in string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	// Connect standard input, output and error.
	var stdin, stdout, stderr bytes.Buffer
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Connect standard input and sent data to it.
	stdin.Write([]byte(in))

	// Run the command.
	if err := cmd.Run(); err != nil {
		if stderr := stderr.Bytes(); len(stderr) > 0 {
			return "error\n" + string(stderr), nil
		}
		if stdout := stdout.Bytes(); len(stdout) > 0 {
			return "error\n" + string(stdout), nil
		}
		return "", err
	}

	return "checked\n" + string(stdout.Bytes()), nil
}

////////////////////////////////////////////////////////////////////////////////
// Feedback

func feedback() error {
	var grading Grading

	// Load task id from file.
	err := loadTaskId(&grading.Tid)
	if err != nil {
		return err
	}

	// Read and parse test configuration.
	var config TestConfig
	if err := readTestConfig("/task/config/test.json", &config); err != nil {
		return err
	}

	// Read and parse execution output.
	var output TestOutput
	if err := readTestOutput("/tmp/work/output/res.json", &output); err != nil {
		return err
	}

	// Generate the feedback
	var feedback Feedback
	var stats Stats
	stats.Succeeded = 0
	stats.Total = len(config.Predefined)
	grading.Status = "success"

	for i, test := range config.Predefined {
		result := output.Results[i]
		if result.Status == "checked" && test.Output == result.Output {
			stats.Succeeded++
			continue
		}

		if feedback.Example == nil {
			grading.Status = "failed"
			feedback.Example = &Example{
				Input:    test.Input,
				Expected: test.Output,
				Actual:   result.Output,
			}

			if result.Status == "checked" {
				feedback.Message = test.Message
			}
		}
	}

	// Generate feedback result
	feedback.Stats = &stats
	feedback.Score = float32(stats.Succeeded) / float32(stats.Total)
	grading.Feedback = &feedback
	result, err := json.Marshal(grading)
	if err != nil {
		return err
	}
	fmt.Println(string(result))

	return nil
}

func loadTaskId(tid *string) error {
	content, err := ioutil.ReadFile(workDir + "/tid")
	if err != nil {
		return err
	}
	*tid = string(content)
	return nil
}

func readTestOutput(path string, output *TestOutput) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &output)
}
