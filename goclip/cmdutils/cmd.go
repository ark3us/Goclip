package cmdutils

import (
	"Goclip/goclip/log"
	_ "embed"
	"os/exec"
	"strings"
)

//go:embed completions.sh
var complScript string

func GetCompletions(text string) []string {
	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(complScript + "\nget_completions '" + text + "'")
	out, err := cmd.Output()
	if err != nil {
		log.Error("Error getting bash completions: ", err)
		return nil
	}
	// return strings.Split(string(out), "\n")
	var results []string
	for _, res := range strings.Split(string(out), "\n") {
		if res == "" {
			continue
		}
		if !strings.HasPrefix(res, text) {
			parts := strings.Fields(text)
			if len(parts) == 1 {
				res = text + res
			} else {
				parts = append(parts[:len(parts)-1], res)
				res = strings.Join(parts, " ")
				if res == text {
					continue
				}
			}
		}
		results = append(results, res)
	}
	return results
}

func Exec(command string, inTerminal bool) {
	args := strings.Fields(command)
	if inTerminal {
		args = append(args, ";$SHELL")
		args = append([]string{"x-terminal-emulator", "-e"}, args...)
	}
	cmd := exec.Command("nohup", args...)
	err := cmd.Start()
	if err != nil {
		log.Error("Command error: ", err)
	}
}
