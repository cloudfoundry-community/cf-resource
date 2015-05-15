package resource

import "time"

type Source struct {
	API           string `json:"api"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Organization  string `json:"organization"`
	Space         string `json:"space"`
	SkipCertCheck bool   `json:"skip_cert_check"`
}

type Version struct {
	Timestamp time.Time `json:"timestamp"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
