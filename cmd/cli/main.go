package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("OpenBPL - Open Brand Protection Library")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  openbpl <command>")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  version    Show version information")
		fmt.Println("  help       Show this help message")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("OpenBPL v0.1.0-dev")
	case "help":
		fmt.Println("OpenBPL - Open Brand Protection Library")
		fmt.Println("More detailed help coming soon!")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'openbpl help' for usage information")
		os.Exit(1)
	}
}
