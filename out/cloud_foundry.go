package out

type PAAS interface {
	Login(api string, username string, password string) error
	Target(organization string, space string) error
	PushApp(manifest string) error
}

type CloudFoundry struct {
}

func NewCloudFoundry() *CloudFoundry {
	return &CloudFoundry{}
}
