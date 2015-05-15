package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/in"
)

func main() {
	var request in.Request
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		fatal("reading request from stdin", err)
	}

	timestamp := request.Version.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	response := in.Response{
		Version: resource.Version{
			Timestamp: timestamp,
		},
	}

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		fatal("writing response", err)
	}
}

func fatal(message string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %s\n", message, err)
	os.Exit(1)
}
