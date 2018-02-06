package out

import (
	"os"
	"os/exec"
)

//go:generate counterfeiter . PAAS
type PAAS interface {
	Login(api string, username string, password string, insecure bool) error
	Target(organization string, space string) error
	PushApp(manifest string, path string, currentAppName string, dockerUser string, showLogs bool) error
}

type CloudFoundry struct {
	verbose bool
}

func NewCloudFoundry(verbose bool) *CloudFoundry {
	return &CloudFoundry{verbose}
}

func (cf *CloudFoundry) Login(api string, username string, password string, insecure bool) error {
	args := []string{"api", api}
	if insecure {
		args = append(args, "--skip-ssl-validation")
	}

	err := cf.cf(args...).Run()
	if err != nil {
		return err
	}

	return cf.cf("auth", username, password).Run()
}

func (cf *CloudFoundry) Target(organization string, space string) error {
	return cf.cf("target", "-o", organization, "-s", space).Run()
}

func (cf *CloudFoundry) PushApp(manifest string, path string, currentAppName string, dockerUser string, showLogs bool) error {
	args := []string{}

	if currentAppName == "" {
		args = append(args, "push", "-f", manifest)
	} else {
		args = append(args, "zero-downtime-push", currentAppName, "-f", manifest)
		if showLogs {
			args = append(args, "--show-app-log")
		}
	}

	if dockerUser != "" {
		args = append(args, "--docker-username", dockerUser)
	}

	if path != "" {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return chdir(path, cf.cf(args...).Run)
		}

		// path is a zip file, add it to the args
		args = append(args, "-p", path)
	}

	return cf.cf(args...).Run()
}

func chdir(path string, f func() error) error {
	oldpath, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(path)
	if err != nil {
		return err
	}
	defer os.Chdir(oldpath)

	return f()
}

func (cf *CloudFoundry) cf(args ...string) *exec.Cmd {
	cmd := exec.Command("cf", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CF_COLOR=true", "CF_DIAL_TIMEOUT=30")

	if cf.verbose {
		// we have to set CF_TRACE to direct output directly to stderr due to a known issue in the CF CLI
		// when used together with cli plugins like cf autopilot (as used by cf-resource)
		// see also https://github.com/cloudfoundry/cli/blob/master/README.md#known-issues
		cmd.Env = append(cmd.Env, "CF_TRACE=/dev/stderr")
	}

	return cmd
}
