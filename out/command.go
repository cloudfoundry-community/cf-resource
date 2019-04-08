package out

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	"os"

	"github.com/concourse/cf-resource"
)

const CfDockerPassword = "CF_DOCKER_PASSWORD"

type Command struct {
	paas PAAS
}

func NewCommand(paas PAAS) *Command {
	return &Command{
		paas: paas,
	}
}

func (command *Command) Run(request Request) (Response, error) {

	var credentials Credentials

	if request.Params.CredentialsFile != "" {
		rawCredentialsFile, err := ioutil.ReadFile(request.Params.CredentialsFile)
		if err != nil {
			return Response{}, err
		}

		if err := json.NewDecoder(bytes.NewReader(rawCredentialsFile)).Decode(&credentials); err != nil {
			return Response{}, err
		}
	} else {
		credentials = Credentials{Password: request.Source.Password, Username:request.Source.Username}
	}

	err := command.paas.Login(
		request.Source.API,
		credentials.Username,
		credentials.Password,
		request.Source.ClientID,
		request.Source.ClientSecret,
		request.Source.SkipCertCheck,
	)
	if err != nil {
		return Response{}, err
	}

	err = command.paas.Target(
		request.Source.Organization,
		request.Source.Space,
	)
	if err != nil {
		return Response{}, err
	}

	if err := command.setEnvironmentVariables(request); err != nil {
		return Response{}, err
	}

	if request.Params.DockerPassword != "" {
		os.Setenv(CfDockerPassword, request.Params.DockerPassword)
	}

	err = command.paas.PushApp(
		request.Params.ManifestPath,
		request.Params.Path,
		request.Params.CurrentAppName,
		request.Params.Vars,
		request.Params.VarsFiles,
		request.Params.DockerUsername,
		request.Params.ShowAppLog,
		request.Params.NoStart,
	)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Version: resource.Version{
			Timestamp: time.Now(),
		},
		Metadata: []resource.MetadataPair{
			{
				Name:  "organization",
				Value: request.Source.Organization,
			},
			{
				Name:  "space",
				Value: request.Source.Space,
			},
		},
	}, nil
}

func (command *Command) setEnvironmentVariables(request Request) error {
	if len(request.Params.EnvironmentVariables) == 0 {
		return nil
	}

	manifest, err := NewManifest(request.Params.ManifestPath)
	if err != nil {
		return err
	}

	for key, value := range request.Params.EnvironmentVariables {
		manifest.AddEnvironmentVariable(key, value)
	}

	err = manifest.Save(request.Params.ManifestPath)
	if err != nil {
		return err
	}

	return nil
}
