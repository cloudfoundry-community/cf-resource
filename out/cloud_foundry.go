package out

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/concourse/cf-resource/out/rewind"
)

//go:generate counterfeiter . PAAS
type PAAS interface {
	Login(api string, username string, password string, clientID string, clientSecret string, insecure bool) error
	Target(organization string, space string) error
	PushApp(manifest string, path string, currentAppName string, vars map[string]interface{}, varsFiles []string, dockerUser string, showLogs bool, noStart bool) error
}

type CloudFoundry struct {
	verbose bool
}

func NewCloudFoundry(verbose bool) *CloudFoundry {
	return &CloudFoundry{verbose}
}

func (cf *CloudFoundry) Login(api string, username string, password string, clientID string, clientSecret string, insecure bool) error {
	args := []string{"api", api}
	if insecure {
		args = append(args, "--skip-ssl-validation")
	}

	err := cf.cf(args...).Run()
	if err != nil {
		return err
	}

	if clientID != "" && clientSecret != "" {
		return cf.cf("auth", "--client-credentials", clientID, clientSecret).Run()
	}
	return cf.cf("auth", username, password).Run()
}

func (cf *CloudFoundry) Target(organization string, space string) error {
	return cf.cf("target", "-o", organization, "-s", space).Run()
}

func (cf *CloudFoundry) PushApp(
	manifest string,
	path string,
	currentAppName string,
	vars map[string]interface{},
	varsFiles []string,
	dockerUser string,
	showLogs bool,
	noStart bool,
) error {

	if currentAppName == "" {
		return cf.simplePush(manifest, path, currentAppName, vars, varsFiles, dockerUser, noStart)
	} else {
		pushFunction := func() error {
			return cf.simplePush(manifest, path, currentAppName, vars, varsFiles, dockerUser, noStart)
		}

		return rewind.Actions{
			Actions:              cf.zeroDowntimeActions(currentAppName, pushFunction, showLogs),
			RewindFailureMessage: "Oh no. Something's gone wrong. I've tried to roll back but you should check to see if everything is OK.",
		}.Execute()
	}
}

func (cf *CloudFoundry) zeroDowntimeActions(
	currentAppName string,
	pushFunction func() error,
	showLogs bool,
) []rewind.Action {

	venerableAppName := fmt.Sprintf("%s-venerable", currentAppName)

	return []rewind.Action{
		{
			Forward: cf.cf("rename", currentAppName, venerableAppName).Run,
		},
		{
			Forward: pushFunction,
			ReversePrevious: func() error {
				if showLogs {
					_ = cf.cf("logs", "--recent", currentAppName).Run()
				}
				_ = cf.cf("delete", "-f", currentAppName).Run()
				return cf.cf("rename", venerableAppName, currentAppName).Run()
			},
		},
		{
			Forward: cf.cf("delete", "-f", venerableAppName).Run,
		},
	}
}

func (cf *CloudFoundry) simplePush(
	manifest string,
	path string,
	currentAppName string,
	vars map[string]interface{},
	varsFiles []string,
	dockerUser string,
	noStart bool,
) error {

	args := []string{"push"}

	if currentAppName != "" {
		args = append(args, currentAppName)
	}

	args = append(args, "-f", manifest)

	if noStart {
		args = append(args, "--no-start")
	}

	for name, value := range vars {
		args = append(args, "--var", fmt.Sprintf("%s=%s", name, value))
	}

	for _, varsFile := range varsFiles {
		args = append(args, "--vars-file", varsFile)
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
		cmd.Env = append(cmd.Env, "CF_TRACE=true")
	}

	return cmd
}
