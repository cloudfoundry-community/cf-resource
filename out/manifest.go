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

func (manifest *Manifest) EnvironmentVariables() []map[interface{}]interface{} {
	apps, hasApps := manifest.data["applications"].([]interface{})
	if !hasApps {
		return []map[interface{}]interface{}{}
	}
	appEnvVars := make([]map[interface{}]interface{}, len(apps))
	for appIdx := range apps {
		app, isApp := apps[appIdx].(map[interface{}]interface{})
		if !isApp {
			continue
		}
		envVars, hasEnvVars := app["env"].(map[interface{}]interface{})
		if !hasEnvVars {
			envVars = make(map[interface{}]interface{})
			app["env"] = envVars
		}
		appEnvVars[appIdx] = envVars
	}
	return appEnvVars
}

func (manifest *Manifest) AddEnvironmentVariable(name, value string) {
	appEnvVars := manifest.EnvironmentVariables()
	for appIdx := range appEnvVars {
		if appEnvVars[appIdx] != nil {
			appEnvVars[appIdx][name] = value
		}
	}
}

func (manifest *Manifest) Save(manifestPath string) error {
	data, err := yaml.Marshal(manifest.data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(manifestPath, data, 0644)
}
