package cmdutils

import (
	"Goclip/db"
	"Goclip/goclip/log"
	_ "embed"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

const maxCompletions = 500

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
	completions := strings.Split(string(out), "\n")
	// log.Info("Got completions: ", len(completions))
	if len(completions) > maxCompletions {
		completions = completions[:maxCompletions]
	}
	var results []string
	for _, res := range completions {
		if res == "" {
			continue
		}
		// log.Info("Completion: ", res)
		if !strings.HasPrefix(res, text) {
			parts := strings.Fields(text)
			if len(parts) == 1 || strings.HasSuffix(text, " ") {
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
	log.Info("Executing: ", strings.Join(args, " "))
	cmd := exec.Command("nohup", args...)
	err := cmd.Start()
	if err != nil {
		log.Error("Command error: ", err)
	}
}

func ExecEntry(entry *db.ClipboardEntry) {
	tmpFile := "gocliptmp*"
	if entry.IsText() {
		tmpFile += ".txt"
	} else if entry.IsImage() {
		tmpFile += ".png"
	}
	file, err := ioutil.TempFile("/tmp", tmpFile)
	if err != nil {
		log.Warning("Error creating temp file: ", err)
		return
	}
	/*
		defer func() {
			go func() {
				time.Sleep(time.Second)
				log.Info("Removing temp file: ", file.Name())
				if err := os.Remove(file.Name()); err != nil {
					log.Warning("Error removing temp file: ", err)
				}
			}()
		}()
	*/

	if _, err := file.Write(entry.Data); err != nil {
		log.Warning("Error writing to temp file: ", err)
		return
	}
	Exec("xdg-open "+filepath.Join(file.Name()), false)
}
