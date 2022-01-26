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
	"path/filepath"
	"strings"
)

var dbFile = "~/goclip_db"

const (
	argClipboard = "clipboard"
	argApps      = "apps"
)

type GoclipListener struct {
	db           db.GoclipDB
	clipLauncher ui.GoclipLauncher
	appLauncher  ui.GoclipLauncher
	cmdLauncher  ui.GoclipLauncher
}

func NewGoclipListener(goclipDB db.GoclipDB, clipLauncher ui.GoclipLauncher, appLauncher ui.GoclipLauncher, cmdLauncher ui.GoclipLauncher) *GoclipListener {
	return &GoclipListener{
		db:           goclipDB,
		clipLauncher: clipLauncher,
		appLauncher:  appLauncher,
		cmdLauncher:  cmdLauncher,
	}
}

func (s *GoclipListener) Start() {
	go s.startHotkeyListener()
}

func (s *GoclipListener) startHotkeyListener() {
	sets, err := s.db.GetSettings()
	if err != nil {
		sets = db.DefaultSettings()
	}
	hook.Keycode["win"] = 125
	hook.Register(hook.KeyDown, []string{sets.ClipboardKey, sets.ClipboardModKey}, func(event hook.Event) {
		s.clipLauncher.ShowEntries()

	})
	hook.Register(hook.KeyDown, []string{sets.AppsKey, sets.AppsModKey}, func(event hook.Event) {
		s.appLauncher.ShowEntries()
	})
	hook.Register(hook.KeyDown, []string{sets.CmdKey, sets.CmdModKey}, func(event hook.Event) {
		s.cmdLauncher.ShowEntries()
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
	clipLauncher := launcher.NewClipboardLauncher(goclipDB, goclipCb)
	appLauncher := launcher.NewAppsLauncher(goclipDB)
	cmdLauncher := launcher.NewCmdLauncher()
	log.Info("Starting listener")
	goclipSets := settings.New(goclipDB)
	goclipListener := NewGoclipListener(goclipDB, clipLauncher, appLauncher, cmdLauncher)
	goclipCb.Start()
	goclipListener.Start()
	goclipDB.SetRefreshCallback(appLauncher.RedrawApps)
	go goclipDB.RefreshApps()
	goclipSets.Run()
}
