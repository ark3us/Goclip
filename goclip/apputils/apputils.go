package apputils

import (
	"Goclip/db"
	"Goclip/goclip/log"
	"Goclip/goclip/shellutils"
	"bufio"
	"io"
	"io/fs"
	"io/ioutil"
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

type DesktopFileParser struct {
	icons map[string]string
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

func stripExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func (s *DesktopFileParser) ListIcons() map[string]string {
	if s.icons != nil {
		return s.icons
	}
	s.icons = map[string]string{}
	pathEnv := os.Getenv("XDG_DATA_DIRS")
	paths := strings.Split(pathEnv, ":")
	for _, root := range paths {
		root = filepath.Join(root, "icons")
		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".png") {
				key := stripExt(info.Name())
				if _, found := s.icons[key]; !found {
					s.icons[stripExt(info.Name())] = path
				}
			}
			return nil
		})
		if err == io.EOF {
			break
		}
	}
	return s.icons
}

func (s *DesktopFileParser) findIcon(name string, defaultVal string) string {
	if name == "" {
		return defaultVal
	}
	if _, err := os.Stat(name); err == nil {
		return name
	}
	if s.icons == nil {
		s.ListIcons()
	}
	if _, found := s.icons[name]; found {
		return s.icons[name]
	}
	log.Info("Icon not found:", name)
	return defaultVal
}

func (s *DesktopFileParser) ParseDesktopFile(fn string) ([]*db.AppEntry, error) {
	file, err := os.Open(fn)
	if err != nil {
		log.Error("Cannot open file: ", err)
		return nil, err
	}
	var entries []*db.AppEntry
	var entry *db.AppEntry
	scanner := bufio.NewScanner(file)
	fileIcon := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") {
			if entry != nil && entry.Exec != "" {
				entry.Icon = s.findIcon(entry.Icon, fileIcon)
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
			if fileIcon == "" {
				fileIcon = s.findIcon(entry.Icon, "")
			}
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
		entry.Icon = s.findIcon(entry.Icon, fileIcon)
		entries = append(entries, entry)
	}
	return entries, nil
}

type AppManager struct {
	db db.GoclipDB
}

func NewAppManager(myDb db.GoclipDB) *AppManager {
	return &AppManager{db: myDb}
}

func (s *AppManager) LoadApps() {
	parser := DesktopFileParser{}
	pathEnv := os.Getenv("XDG_DATA_DIRS")
	paths := strings.Split(pathEnv, ":")
	var allEntries []*db.AppEntry
	for _, path := range paths {
		path = filepath.Join(path, "applications")
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Warning("Cannot read dir content: ", err)
			continue
		}
		n := 0
		for _, finfo := range files {
			ffile := filepath.Join(path, finfo.Name())
			entries, err := parser.ParseDesktopFile(ffile)
			if err != nil {
				continue
			}
			allEntries = append(allEntries, entries...)
			n += len(entries)
		}
		log.Info(path, ": ", n)
	}
	if err := s.db.AddAppEntries(allEntries); err != nil {
		log.Error("Error saving app entries: ", err)
	}
}

func (s *AppManager) GetApps() []*db.AppEntry {
	return s.db.GetAppEntries()
}

func (s *AppManager) ExecEntry(entry *db.AppEntry) {
	shellutils.Exec(entry.Exec, entry.Terminal)
	s.db.UpdateAppEntry(entry)
}
