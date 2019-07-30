// Pythia library for unit testing-based tasks
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
	"flag"
	"fmt"
	"os"
)

var fcts = map[string]func(){
	"preprocess": preprocess,
	"generate":   generate,
	"feedback":   feedback,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Subcommand is required (preprocess, generate or feedback).")
		os.Exit(1)
	}
	if handle, ok := fcts[os.Args[1]]; ok {
		handle()
	}
}

func preprocess() {
	preprocessCmd := flag.NewFlagSet("preprocess", flag.ExitOnError)
	preprocessCmd.Parse(os.Args[2:])
	fmt.Println("preprocess")
}

func generate() {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCmd.Parse(os.Args[2:])
	fmt.Println("generate")
}

func feedback() {
	feedbackCmd := flag.NewFlagSet("feedback", flag.ExitOnError)
	feedbackCmd.Parse(os.Args[2:])
	fmt.Println("feedback")
}
