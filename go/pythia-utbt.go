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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type TaskData struct {
	Tid    string            `json:"tid"`
	Fields map[string]string `json:"fields"`
}

type SpecConfig struct {
	Name string `json:"name"`
	Args []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"args"`
}

type TestConfig struct {
	Predefined []struct {
		Data     string            `json:"data"`
		Feedback map[string]string `json:"feedback,omitempty"`
	} `json:"predefined,omitempty"`
	Random struct {
		N    int      `json:"n"`
		Args []string `json:"args"`
	} `json:"random"`
}

type Example struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type Stats struct {
	Succeeded int `json:"succeeded"`
	Total     int `json:"total"`
}

type Feedback struct {
	Message string   `json:"message,omitempty"`
	Example *Example `json:"example,omitempty"`
	Stats   *Stats   `json:"stats,omitempty"`
	Score   float32  `json:"score,omitempty"`
}

type Grading struct {
	Tid      string    `json:"tid"`
	Status   string    `json:"status"`
	Feedback *Feedback `json:"feedback,omitempty"`
}

var fcts = map[string]func(){
	"preprocess": preprocess,
	"generate":   generate,
	"execute":    execute,
	"feedback":   feedback,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Subcommand is required (preprocess, generate, execute or feedback).")
		os.Exit(1)
	}
	if handle, ok := fcts[os.Args[1]]; ok {
		handle()
	}
}

func preprocess() {
	// Setup working directory
	dirName := "/tmp/work"
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err := os.MkdirAll(dirName, 0777); err != nil {
			panic(err)
		}
	}

	// Create directories for input and output data for unit tests
	if err := os.MkdirAll(dirName+"/input", 0777); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(dirName+"/output", 0777); err != nil {
		panic(err)
	}

	// Read input data
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	var data TaskData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		panic(err)
	}

	// Fill skeleton files
	fillSkeletonFiles("/tmp/work", data.Fields)

	// Save task id
	if err := ioutil.WriteFile(dirName+"/tid", []byte(data.Tid), 0444); err != nil {
		panic(err)
	}
}

func fillSkeletonFiles(dest string, fields map[string]string) {
	if err := filepath.Walk("/task/skeleton",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsDir() {
				destDir := dest + filepath.Dir(path)[14:]
				destFile := destDir + "/" + filepath.Base(path)

				if err := os.MkdirAll(destDir, 0777); err != nil {
					panic(err)
				}
				content, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				for key, value := range fields {
					regex, _ := regexp.Compile("@([^@]*)@" + key + "@([^@]*)@")
					matches := regex.FindAllStringSubmatch(string(content), -1)
					for _, match := range matches {
						lines := strings.Split(value, "\n")
						rep := ""
						for _, line := range lines {
							rep += match[1] + line + match[2] + "\n"
						}
						content = []byte(strings.ReplaceAll(string(content), "@"+match[1]+"@"+key+"@"+match[2]+"@", rep))
					}
				}

				if err := ioutil.WriteFile(destFile, content, 0774); err != nil {
					panic(err)
				}
			}
			return nil
		},
	); err != nil {
		panic(err)
	}
}

func generate() {
	// Read test configuration
	content, err := ioutil.ReadFile("/task/config/test.json")
	if err != nil {
		panic(err)
	}

	var config TestConfig
	if err := json.Unmarshal(content, &config); err != nil {
		panic(err)
	}

	// Create test data file
	file, err := os.Create("/tmp/work/input/data.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := os.Chmod("/tmp/work/input/data.csv", 0774); err != nil {
		panic(err)
	}
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	// Generate test data
	if config.Predefined != nil {
		for _, data := range config.Predefined {
			inputs := strings.Split(data.Data[1:len(data.Data)-1], ",")
			for i := range inputs {
				inputs[i] = strings.TrimSpace(inputs[i])
			}
			writer.Write(inputs)
		}
	}
	/*
	       # Create an array of generators as specified by configuration
	       # and write random tests to the specified file if any
	       if 'random' in config:
	           random = config['random']
	           generator = ArrayGenerator([RandomGenerator.build(descr) for descr in random['args']])
	           for i in range(random['n']):
	               writer.writerow(generator.generate())
	   os.chmod(filedest, stat.S_IRWXU | stat.S_IRWXG | stat.S_IROTH)

	*/
}

func execute() {
	// Execute the code from the learner
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func feedback() {
	var grading Grading
	var content []byte
	var err error

	// Retrieve task id
	content, err = ioutil.ReadFile("/tmp/work/tid")
	if err != nil {
		panic(err)
	}
	grading.Tid = string(content)

	// Check if there are any standard error
	content, err = ioutil.ReadFile("/tmp/work/output/out.err")
	if err == nil {
		grading.Status = "failed"
		grading.Feedback = &Feedback{
			Message: string(content),
		}
		result, err := json.Marshal(grading)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(result))
		os.Exit(0)
	}

	// Check if there are any standard output

	// Read author solution
	content, err = ioutil.ReadFile("/task/config/solution")
	if err != nil {
		panic(err)
	}
	solution := string(content)

	// Fill skeleton files
	solutionDir := "/tmp/work/solution"
	if err := os.MkdirAll(solutionDir, 0777); err != nil {
		panic(err)
	}
	fillSkeletonFiles(solutionDir, map[string]string{"f1": solution})

	// Execute the code from the author
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Generate the feedback
	var feedback Feedback
	var stats Stats
	stats.Succeeded = 0
	stats.Total = 0
	grading.Status = "success"

	results, err := ReadLines("/tmp/work/output/data.res")
	if err != nil {
		panic(err)
	}
	solutions, err := ReadLines("/tmp/work/output/solution.res")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("/tmp/work/input/data.csv")
	if err != nil {
		panic(err)
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
			panic(err)
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
		panic(err)
	}
	fmt.Println(string(result))
}

func ReadLines(path string) (lines []string, err error) {
	if content, err := ioutil.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimSpace(string(content)), "\n")
	}
	return lines, err
}
