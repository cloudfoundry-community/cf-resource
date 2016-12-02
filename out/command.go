package out

import (
	"time"

	"github.com/concourse/cf-resource"
)

type Command struct {
	paas PAAS
}

func NewCommand(paas PAAS) *Command {
	return &Command{
		paas: paas,
	}
}

func (command *Command) Run(request Request) (Response, error) {
	err := command.paas.Login(
		request.Source.API,
		request.Source.Username,
		request.Source.Password,
		request.Source.SkipCertCheck,
	)
	if err != nil {
		return Response{}, err
	}

	organization := request.Source.Organization
	space := request.Source.Space
	if request.Params.Organization != "" {
		organization = request.Params.Organization
	}

	if request.Params.Space != "" {
		space = request.Params.Space
	}

	err = command.paas.Target(
		organization,
		space,
	)
	if err != nil {
		return Response{}, err
	}

	if err := command.setEnvironmentVariables(request); err != nil {
		return Response{}, err
	}

	err = command.paas.PushApp(
		request.Params.ManifestPath,
		request.Params.Path,
		request.Params.CurrentAppName,
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
				Value: organization,
			},
			{
				Name:  "space",
				Value: space,
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
