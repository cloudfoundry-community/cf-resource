package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/concourse/cf-resource/out"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <sources directory>\n", os.Args[0])
		os.Exit(1)
	}

	cloudFoundry := out.NewCloudFoundry()
	command := out.NewCommand(cloudFoundry)

	var request out.Request
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		fatal("reading request from stdin", err)
	}

	// make it an absolute path
	request.Params.ManifestPath = filepath.Join(os.Args[1], request.Params.ManifestPath)

	if request.Params.Path != "" {
		request.Params.Path = filepath.Join(os.Args[1], request.Params.Path)
	}

	response, err := command.Run(request)
	if err != nil {
		fatal("running command", err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		fatal("writing response to stdout", err)
	}
}

func fatal(message string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %s\n", message, err)
	os.Exit(1)
}
