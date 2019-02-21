package zdt

import (
	"fmt"
	"os/exec"
)

func Push(
	cf func(args ...string) *exec.Cmd,
	currentAppName string,
	pushFunction func() error,
	showLogs bool,
) error {

	venerableAppName := fmt.Sprintf("%s-venerable", currentAppName)

	actions := []Action{
		{
			Forward: cf("rename", currentAppName, venerableAppName).Run,
		},
		{
			Forward: pushFunction,
			ReversePrevious: func() error {
				if showLogs {
					_ = cf("logs", currentAppName, "--recent").Run()
				}
				_ = cf("delete", "-f", currentAppName).Run()
				return cf("rename", venerableAppName, currentAppName).Run()
			},
		},
		{
			Forward: cf("delete", "-f", venerableAppName).Run,
		},
	}

	return Actions{
		Actions:              actions,
		RewindFailureMessage: "Oh no. Something's gone wrong. I've tried to roll back but you should check to see if everything is OK.",
	}.Execute()
}
