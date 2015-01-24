package out

import "time"

type Source struct {
	API           string `json:"api"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Organization  string `json:"organization"`
	Space         string `json:"space"`
	SkipCertCheck bool   `json:"skip_cert_check"`
}

type Request struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type Params struct {
	ManifestPath   string `json:"manifest"`
	Path           string `json:"path"`
	CurrentAppName string `json:"current_app_name"`
}

type Response struct {
	Version  Version        `json:"version"`
	Metadata []MetadataPair `json:"metadata"`
}

type Version struct {
	Timestamp time.Time `json:"timestamp"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
