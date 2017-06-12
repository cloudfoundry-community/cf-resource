package out

import (
	"strings"
	"os"
	"fmt"
)

type CfEnvironment struct {
	env map[string]string
}

func NewCfEnvironment() *CfEnvironment {
	env := make(map[string]string)
	env["CF_COLOR"]="true"

	cfe := &CfEnvironment{env}

	return cfe
}

func NewCfEnvironmentFromOS() *CfEnvironment {
	cfe := NewCfEnvironment()

	osEnvironment := SplitKeyValueStringArrayInToMap(os.Environ())
	cfe.AddEnvironmentVariable(osEnvironment)

	return cfe
}

func SplitKeyValueStringArrayInToMap(data []string) map[string]interface{} {
	items := make(map[string]interface{})
	for _, item := range data {
		key, val := SplitKeyValueString(item)
		items[key] = val
	}
	return items
}

func SplitKeyValueString(item string)(key, val string) {
	splits := strings.SplitN(item, "=", 2)
	key = splits[0]
	val = splits[1]
	return
}


func (cfe *CfEnvironment) addEnvironmentVariable(key, value string) {
	cfe.env[key] = value
}

func (cfe *CfEnvironment) Environment() []string {
	var commandEnvironment []string

	for k, v := range cfe.env {
		commandEnvironment = append(commandEnvironment, k+"="+v)
	}
	return commandEnvironment
}

func (cfe *CfEnvironment) AddEnvironmentVariable(switchMap map[string]interface{}) {
	for k, v := range switchMap {
		vString := fmt.Sprintf("%v", v)
		cfe.env[k] = vString
	}
}
