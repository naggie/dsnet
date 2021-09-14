package utils

import (
	"fmt"
	"os/exec"
)

func ShellOut(command string, name string) error {
	if command != "" {
		shell := exec.Command("/bin/sh", "-c", command)
		err := shell.Run()
		if err != nil {
			return fmt.Errorf("failed to execute(%s - `%s`): %s ", name, command, err)
		}
	}
	return nil
}
