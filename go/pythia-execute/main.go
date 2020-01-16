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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pythia-project/libs/go/pythia/utils"
)

func main() {
	var execResult utils.ExecutionResult

	// Parse arguments.
	fileName := flag.String("filename", "", "Program source code file name.")
	compileCmd := flag.String("compile", "", "Command to compile the program.")
	executeCmd := flag.String("execute", "", "Command to execute the program.")
	flag.Parse()

	// Setup working directory.
	if err := utils.SetupWorkDir(); err != nil {
		log.Fatalf("Error while creating working directory: %s.", err)
	}

	// Read input data.
	input, err := utils.ReadStdIn()
	if err != nil {
		log.Fatalf("Error while reading stdin: %s.", err)
	}

	// Create source code file.
	srcFile := fmt.Sprintf("%s/%s", utils.WORKDIR, *fileName)
	if err := ioutil.WriteFile(srcFile, input, 0774); err != nil {
		log.Fatalf("Error while creating source code file: %s.", err)
	}

	// Compile and execute program.
	if *compileCmd != "" {
		execResult = utils.Execute(compileCmd, "")
	}
	if *executeCmd != "" && execResult.ReturnCode == 0 {
		execResult = utils.Execute(executeCmd, "")
	}

	// Generate JSON execution result.
	result, err := json.Marshal(execResult)
	if err != nil {
		log.Fatalf("Error while generating JSON output: %s.", err)
	}
	fmt.Println(string(result))
	os.Exit(0)
}
