package out

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Manifest struct {
	data map[interface{}]interface{}
}

func NewManifest(manifestPath string) (Manifest, error) {
	yamlData, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(yamlData, &manifest.data)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func (manifest *Manifest) EnvironmentVariables() map[interface{}]interface{} {
	envVars, hasEnvVars := manifest.data["env"].(map[interface{}]interface{})
	if !hasEnvVars {
		envVars = make(map[interface{}]interface{})
		manifest.data["env"] = envVars
	}

	return envVars
}

func (manifest *Manifest) AddEnvironmentVariable(name, value string) {
	manifest.EnvironmentVariables()[name] = value
}

func (manifest *Manifest) Save(manifestPath string) error {
	data, err := yaml.Marshal(manifest.data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(manifestPath, data, 0644)
}
