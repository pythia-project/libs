package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	preprocessCmd := flag.NewFlagSet("preprocess", flag.ExitOnError)
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	feedbackCmd := flag.NewFlagSet("feedback", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "preprocess":
		preprocessCmd.Parse(os.Args[2:])
		fmt.Println("preprocess")
	case "generate":
		generateCmd.Parse(os.Args[2:])
		fmt.Println("generate")
	case "feedback":
		feedbackCmd.Parse(os.Args[2:])
		fmt.Println("feedback")
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
