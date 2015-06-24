package out

import "github.com/concourse/cf-resource"

type Request struct {
	Source resource.Source `json:"source"`
	Params Params          `json:"params"`
}

type Params struct {
	ManifestPath         string            `json:"manifest"`
	Path                 string            `json:"path"`
	CurrentAppName       string            `json:"current_app_name"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

type Response struct {
	Version  resource.Version        `json:"version"`
	Metadata []resource.MetadataPair `json:"metadata"`
}
