package out

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/concourse/cf-resource/out/zdt"
)

//go:generate counterfeiter . PAAS
type PAAS interface {
	ApplyManifest(currentAppName string, manifest string) error
	Login(api string, username string, password string, clientID string, clientSecret string, insecure bool) error
	Target(organization string, space string) error
	PushApp(manifest string, path string, currentAppName string, vars map[string]interface{}, varsFiles []string, dockerUser string, showLogs bool, noStart bool) error
	PushAppWithRollingDeployment(path string, currentAppName string, dockerUser string, showLogs bool, noStart bool, manifest string) error
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

func (cf *CloudFoundry) PushAppWithRollingDeployment(
	path string,
	currentAppName string,
	dockerUser string,
	showLogs bool,
	noStart bool,
	manifest string,
) error {
	cf.cf("version").Run()

	if manifest != "" {
		cf.ApplyManifest(currentAppName, manifest)
	}
	return cf.simpleRollingDeploymentPush(path, currentAppName, dockerUser, noStart)
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
	if zdt.CanPush(cf.cf, currentAppName) {
		pushFunction := func() error {
			return cf.simplePush(manifest, path, currentAppName, vars, varsFiles, dockerUser, noStart)
		}
		return zdt.Push(cf.cf, currentAppName, pushFunction, showLogs)
	} else {
		return cf.simplePush(manifest, path, currentAppName, vars, varsFiles, dockerUser, noStart)
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
			args = append(args, "-p", ".")
			return chdir(path, cf.cf(args...).Run)
		}

		// path is a zip file, add it to the args
		args = append(args, "-p", path)
	}

	return cf.cf(args...).Run()
}

func (cf *CloudFoundry) ApplyManifest(currentAppName, manifest string) error {
	return cf.cf("v3-apply-manifest", currentAppName, "-f", manifest).Run()
}

func (cf *CloudFoundry) simpleRollingDeploymentPush(
	path string,
	currentAppName string,
	dockerUser string,
	noStart bool,
) error {
	args := []string{"v3-zdt-push", "--wait-for-deploy-complete"}

	if currentAppName != "" {
		args = append(args, currentAppName)
	}

	if noStart {
		args = append(args, "--no-start")
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
			args = append(args, "-p", ".")
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
