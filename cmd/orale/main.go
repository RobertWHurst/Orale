package main

import (
	"fmt"
	"os"

	"github.com/RobertWHurst/orale/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Printf("orale error: %s\n", err)
		os.Exit(1)
	}
}
