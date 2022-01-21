package main

import (
	"Goclip/clipboard"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/goclip/log"
	"Goclip/ui"
	"Goclip/ui/gtk/launcher"
	"Goclip/ui/gtk/settings"
	hook "github.com/robotn/gohook"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var dbFile = "~/goclip_db"

const (
	argClipboard = "clipboard"
	argApps      = "apps"
)

type GoclipListener struct {
	app      string
	db       db.GoclipDB
	settings ui.GoclipSettings
}

func NewGoclipListener(goclipApp string, goclipDB db.GoclipDB) *GoclipListener {
	return &GoclipListener{
		app: goclipApp,
		db:  goclipDB,
	}
}

func (s *GoclipListener) Start() {
	go s.startHotkeyListener()
}

func (s *GoclipListener) startLauncher(arg string) {
	cmd := exec.Command(s.app, arg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Error("Error starting launcher: ", err)
	}
}

func (s *GoclipListener) startHotkeyListener() {
	sets, err := s.db.GetSettings()
	if err != nil {
		sets = db.DefaultSettings()
	}
	hook.Register(hook.KeyDown, []string{sets.ClipboardKey, sets.ClipboardModKey}, func(event hook.Event) {
		go s.startLauncher(argClipboard)
	})
	hook.Register(hook.KeyDown, []string{sets.AppsKey, sets.AppsModKey}, func(event hook.Event) {
		go s.startLauncher(argApps)
	})
	start := hook.Start()
	<-hook.Process(start)
}

func main() {
	// log.Debug = true
	if strings.HasPrefix(dbFile, "~/") {
		dirname, _ := os.UserHomeDir()
		dbFile = filepath.Join(dirname, dbFile[2:])
	}
	goclipDB := storm.New(dbFile)
	goclipCb := clipboard.New(goclipDB)
	if len(os.Args) > 1 && os.Args[1] == argClipboard {
		log.Info("Staring clipboard launcher")
		goclipApp := launcher.NewClipboardLauncher(goclipDB, goclipCb)
		goclipApp.Run()
	} else if len(os.Args) > 1 && os.Args[1] == argApps {
		log.Info("Staring app launcher")
		goclipApp := launcher.NewAppsLauncher(goclipDB)
		goclipApp.Run()
	} else {
		log.Info("Starting listener")
		ex, err := os.Executable()
		if err != nil {
			log.Fatal("Error getting executable:", err)
		}
		goclipApp := settings.New(goclipDB)
		goclipListener := NewGoclipListener(ex, goclipDB)
		goclipCb.Start()
		goclipListener.Start()
		go goclipDB.RefreshApps()
		goclipApp.Run()
	}
}
