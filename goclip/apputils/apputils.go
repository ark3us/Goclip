package apputils

import (
	"Goclip/db"
	"Goclip/goclip/log"
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func ATime(name string) (atime time.Time) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	return
}

func removeExecFieldCodes(value string) string {
	return strings.NewReplacer(
		"%f", "",
		"%F", "",
		"%u", "",
		"%U", "",
		"%d", "",
		"%D", "",
		"%n", "",
		"%N", "",
		"%i", "",
		"%c", "",
		"%k", "",
		"%v", "",
		"%m", "").Replace(value)
}

func getDesktopFileValue(line string, prefix string) string {
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	parts := strings.Split(line, prefix)
	if len(parts) < 2 {
		return ""
	}
	return removeExecFieldCodes(strings.Join(parts[1:], prefix))
}

func findIcon(name string) string {
	if _, err := os.Stat(name); err == nil {
		return name
	}
	pathEnv := os.Getenv("XDG_DATA_DIRS")
	paths := strings.Split(pathEnv, ":")
	out := "default.png"
	for _, root := range paths {
		log.Info("Looking for icon ", name, " in ", root)
		root = filepath.Join(root, "icons")
		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == name+".png" {
				out = path
				return io.EOF
			}
			return nil
		})
		if err == io.EOF {
			err = nil
		}
		if out != "" {
			break
		}
	}
	return out
}

func ParseDesktopFile(fn string) ([]*db.AppEntry, error) {
	file, err := os.Open(fn)
	if err != nil {
		log.Error("Cannot open file: ", err)
		return nil, err
	}
	var entries []*db.AppEntry
	var entry *db.AppEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") {
			if entry != nil && entry.Exec != "" {
				entry.Icon = findIcon(entry.Icon)
				entries = append(entries, entry)
			}
			entry = &db.AppEntry{
				AccessTime: ATime(fn),
				File:       fn,
			}
		}

		name := getDesktopFileValue(line, "Name=")
		if name != "" {
			entry.Name = name
		}
		icon := getDesktopFileValue(line, "Icon=")
		if icon != "" {
			entry.Icon = icon
		}
		exec := getDesktopFileValue(line, "Exec=")
		if exec != "" {
			entry.Exec = exec
		}
		term := getDesktopFileValue(line, "Terminal=")
		if term != "" {
			if term == "true" {
				entry.Terminal = true
			}
		}
	}
	if entry.Exec != "" {
		entry.Icon = findIcon(entry.Icon)
		entries = append(entries, entry)
	}
	return entries, nil
}
