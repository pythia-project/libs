// Pythia utilities for unit testing-based tasks
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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pythia-project/libs/go/pythia-utbt/generators"
)

// TaskInput contains the inputs of the learner for the specified task id.
type TaskInput struct {
	Tid    string            `json:"tid"`
	Fields map[string]string `json:"fields"`
}

// TestConfig contains the configuration of the tests for a task.
type TestConfig struct {
	Predefined []struct {
		Data     string            `json:"data"`
		Feedback map[string]string `json:"feedback,omitempty"`
	} `json:"predefined,omitempty"`
	Random struct {
		N    int      `json:"n"`
		Args []string `json:"args"`
	} `json:"random,omitempty"`
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
	Score   float32  `json:"score,omitempty"`
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
	teacherDir = workDir + "/teacher"
)

var fcts = map[string]func() error{
	"preprocess": preprocess,
	"generate":   generate,
	"execute":    execute,
	"feedback":   feedback,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Subcommand is required (preprocess, generate, execute or feedback).")
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
	// Setup working directory and create directories for input/output data for unit tests.
	os.RemoveAll(workDir)
	if err := createDir(0755, workDir, workDir+"/input"); err != nil {
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
// Generate

func generate() error {
	var testInputFile = workDir + "/input/data.csv"

	// Read and parse test configuration.
	var config TestConfig
	if err := readTestConfig("/task/config/test.json", &config); err != nil {
		return err
	}

	// Create test inputs CSV file.
	file, err := os.Create(testInputFile)
	if err != nil {
		return err
	}
	defer os.Chmod(testInputFile, 0444)
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	// Generate predefined test inputs.
	if config.Predefined != nil {
		for _, data := range config.Predefined {
			writer.Write(parseTestInputs(data.Data))
		}
	}

	// Generate random test inputs.
	if config.Random.N > 0 {
		generators := generators.BuildGenerators(config.Random.Args...)
		for i := 0; i < config.Random.N; i++ {
			writer.Write(generateTestInputs(generators))
		}
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

func parseTestInputs(str string) []string {
	inputs := strings.Split(str[1:len(str)-1], ",")
	for i := range inputs {
		inputs[i] = strings.TrimSpace(inputs[i])
	}

	return inputs
}

func generateTestInputs(gens []generators.RandomGenerator) []string {
	inputs := make([]string, len(gens))
	for i, g := range gens {
		inputs[i] = g.Generate()
	}

	return inputs
}

////////////////////////////////////////////////////////////////////////////////
// Execute

func execute() error {
	if len(os.Args) < 3 {
		return errors.New("Command to execute is missing.")
	}

	// Execute the code from the learner.
	if err := executeCommand(os.Args[2], os.Args[3:]...); err != nil {
		return err
	}

	return nil
}

func executeCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
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

	// Check and handle standard error, if there is any.
	content, err := ioutil.ReadFile(workDir + "/output/out.err")
	if err == nil {
		grading.Status = "failed"
		grading.Feedback = &Feedback{
			Message: string(content),
		}
		if err := printGrading(grading); err != nil {
			return err
		}
		return nil
	}

	// Check and handle standard output, if there is any.

	// Generate the solution.
	if err := executeSolution(); err != nil {
		return err
	}

	// Generate the feedback
	var feedback Feedback
	var stats Stats
	stats.Succeeded = 0
	stats.Total = 0
	grading.Status = "success"

	results, err := readLines(workDir + "/output/data.res")
	if err != nil {
		return err
	}
	solutions, err := readLines(workDir + "/output/solution.res")
	if err != nil {
		return err
	}

	file, err := os.Open(workDir + "/input/data.csv")
	if err != nil {
		return err
	}
	reader := csv.NewReader(file)
	reader.Comma = ';'
	i := -1
	for {
		i++
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		tokens := strings.Split(results[i], ":")
		switch tokens[0] {
		case "checked":
			if tokens[1] == solutions[i] {
				stats.Succeeded++
				continue
			}

			if feedback.Example == nil {
				grading.Status = "failed"
				feedback.Example = &Example{
					Input:    "(" + strings.Join(row, ",") + ")",
					Expected: solutions[i],
					Actual:   tokens[1],
				}
			}
		default:
			grading.Status = "failed"
			if feedback.Message == "" {
				feedback.Message = "An error occurred with your code: " + tokens[0]
				continue
			}
		}
	}
	stats.Total = i

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

func executeSolution() error {
	// Read author solution.
	var solution map[string]string
	if err := readSolution(&solution); err != nil {
		return err
	}

	// Prepare working directory for solution execution.
	os.RemoveAll(teacherDir)
	if err := createDir(0777, teacherDir); err != nil {
		return err
	}

	// Fill skeleton files with author solution.
	if err := fillSkeletonFiles(skeletonDir, teacherDir, solution); err != nil {
		return err
	}

	// Execute the code from the author.
	if err := executeCommand(os.Args[2], os.Args[3:]...); err != nil {
		return err
	}

	return nil
}

func readSolution(solution *map[string]string) error {
	content, err := ioutil.ReadFile("/task/config/solution.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(content, solution)
}

func readLines(path string) (lines []string, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(content)), "\n"), nil
}

func printGrading(grading Grading) error {
	result, err := json.Marshal(grading)
	if err != nil {
		return err
	}
	fmt.Println(string(result))
	return nil
}
