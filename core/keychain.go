package core

import (
	"os/exec"
)

func GetKeychainPassword(account string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-wa", account)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	output = output[:len(output)-1]
	return string(output), nil
}
