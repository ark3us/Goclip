package shellutils

import (
	"Goclip/db"
	"Goclip/goclip/log"
	"bytes"
	_ "embed"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

const maxCompletions = 500
const maxHistory = 1000

//go:embed bash_completions.sh
var bashCompletions string

//go:embed zsh_completions.sh
var zshCompletions string

func ExpandUserDir(path string) (string, error) {
	if !strings.Contains(path, "~/") {
		return "", errors.New("nothing to replace")
	}
	usr, err := user.Current()
	if err != nil {
		log.Error("Cannot get current user: ", err)
		return "", err
	}
	return strings.Replace(path, "~/", usr.HomeDir+"/", -1), nil
}

type ShellManager struct {
	db db.GoclipDB
}

func NewShellManager(myDb db.GoclipDB) *ShellManager {
	return &ShellManager{db: myDb}
}

func (s *ShellManager) LoadHistory() {
	var results []*db.ShellEntry
	var err error
	var data []byte
	r := regexp.MustCompile(`^: \d+:\d+;(.+)$`)

	log.Info("Loading shell history...")

	isZsh := false
	isZshExt := false
	histFile := ""
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		histFile = "~/.zsh_history"
		isZsh = true
	} else if strings.Contains(shell, "bash") {
		histFile = "~/.bash_history"
	} else {
		log.Error("Shell not supported: ", shell)
		return
	}
	if histFile, err = ExpandUserDir(histFile); err != nil {
		return
	}
	if data, err = ioutil.ReadFile(histFile); err != nil {
		log.Error("Error reading history file: ", err)
		return
	}
	if isZsh {
		data = bytes.Replace(data, []byte("\\\n"), []byte(" "), -1)
	}
	lines := strings.Split(string(data), "\n")
	if isZsh && len(lines) > 0 {
		if matches := r.FindStringSubmatch(lines[0]); len(matches) > 0 {
			isZshExt = true
		}
	}
	for i := 0; i < len(lines) && i < maxHistory; i++ {
		line := lines[len(lines)-i-1]
		if isZshExt {
			matches := r.FindStringSubmatch(line)
			if len(matches) == 2 {
				line = matches[1]
			} else {
				continue
			}
		}
		results = append(results, &db.ShellEntry{Cmd: line, IsHistory: true})
	}
	// log.Info(results)
	if err := s.db.AddShellEntries(results); err != nil {
		log.Error("Error saving shell history: ", err)
	}
	log.Info("Loaded history entries: ", len(results))
	return
}

func (s *ShellManager) GetShellCompletions(text string) []*db.ShellEntry {
	var cmd *exec.Cmd
	results, err := s.db.GetShellEntries(text, maxCompletions)
	call := "\nget_completions '" + text + "'"
	shell := os.Getenv("SHELL")
	cmd = exec.Command(shell)
	if strings.Contains(shell, "zsh") {
		cmd.Stdin = strings.NewReader(zshCompletions + call)
	} else if strings.Contains(shell, "bash") {
		cmd.Stdin = strings.NewReader(bashCompletions + call)
	} else {
		return results
	}

	out, err := cmd.Output()
	if err != nil {
		log.Error("Error getting shell completions: ", err)
		return results
	}
	completions := strings.Split(string(out), "\n")
	// log.Info("Got completions: ", len(completions))
	if len(completions) > maxCompletions {
		completions = completions[:maxCompletions]
	}
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
		results = append(results, &db.ShellEntry{Cmd: res, IsShell: true})
	}
	return results
}

func Exec(command string, inTerminal bool) {
	var args []string
	if inTerminal {
		args = append([]string{"x-terminal-emulator", "-x"}, "$SHELL -i -c '"+command+";$SHELL'")
	} else {
		args = strings.Fields(command)
	}
	log.Info("Executing: ", strings.Join(args, " "))
	cmd := exec.Command("nohup", args...)
	err := cmd.Start()
	if err != nil {
		log.Error("Command error: ", err)
	}
}

func OpenEntry(entry *db.ClipboardEntry) {
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
