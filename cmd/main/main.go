package main

import (
	"Goclip/clipboard"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/goclip/log"
	"Goclip/ui"
	"Goclip/ui/gtk/launcher"
	"Goclip/ui/gtk/settings"
	"github.com/gotk3/gotk3/gtk"
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
	app string
	db  db.GoclipDB

	clipLauncher ui.GoclipLauncher
	appLauncher  ui.GoclipLauncher
}

func NewGoclipListener(goclipApp string, goclipDB db.GoclipDB, clipLauncher ui.GoclipLauncher, appLauncher ui.GoclipLauncher) *GoclipListener {
	return &GoclipListener{
		app:          goclipApp,
		db:           goclipDB,
		clipLauncher: clipLauncher,
		appLauncher:  appLauncher,
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
		// go s.startLauncher(argClipboard)
		s.clipLauncher.ShowEntries()

	})
	hook.Register(hook.KeyDown, []string{sets.AppsKey, sets.AppsModKey}, func(event hook.Event) {
		// go s.startLauncher(argApps)
		s.appLauncher.ShowEntries()
	})
	start := hook.Start()
	<-hook.Process(start)
}

func main() {
	gtk.Init(nil)
	// log.Debug = true
	if strings.HasPrefix(dbFile, "~/") {
		dirname, _ := os.UserHomeDir()
		dbFile = filepath.Join(dirname, dbFile[2:])
	}
	goclipDB := storm.New(dbFile)
	goclipCb := clipboard.New(goclipDB)
	clipLauncher := launcher.NewClipboardLauncher(goclipDB, goclipCb)
	appLauncher := launcher.NewAppsLauncher(goclipDB)
	log.Info("Starting listener")
	ex, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable:", err)
	}
	goclipSets := settings.New(goclipDB)
	goclipListener := NewGoclipListener(ex, goclipDB, clipLauncher, appLauncher)
	goclipCb.Start()
	goclipListener.Start()
	go goclipDB.RefreshApps()
	goclipSets.Run()
	gtk.Main()
}
