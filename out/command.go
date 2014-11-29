package out

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

	err = command.paas.PushApp(
		request.Params.ManifestPath,
	)
	if err != nil {
		return Response{}, err
	}

	return Response{}, nil
}
