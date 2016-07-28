package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/tak7iji/consolebeat/beater"
)

func main() {
	err := beat.Run("consolebeat", "0.0.1", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
