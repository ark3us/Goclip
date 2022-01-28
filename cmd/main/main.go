package main

import (
	"Goclip/apputils"
	"Goclip/cliputils"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/log"
	"Goclip/shellutils"
	"Goclip/ui"
	"Goclip/ui/gtk/launcher"
	"Goclip/ui/gtk/settings"
	hook "github.com/robotn/gohook"
	"os"
	"path/filepath"
	"strings"
)

var dbDir = "~/goclip"

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

func HotkeyListener(goclipDB db.GoclipDB, clipLauncher ui.GoclipLauncher, appLauncher ui.GoclipLauncher, cmdLauncher ui.GoclipLauncher) *GoclipListener {
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
	if strings.HasPrefix(dbDir, "~/") {
		dirname, _ := os.UserHomeDir()
		dbDir = filepath.Join(dirname, dbDir[2:])
	}
	goclipDb, err := storm.New(dbDir)
	if err != nil {
		return
	}
	clipManager := cliputils.NewClipboardManager(goclipDb)
	clipManager.StartListener()
	appManager := apputils.NewAppManager(goclipDb)
	shellManager := shellutils.NewShellManager(goclipDb)
	go shellManager.LoadHistory()

	clipLauncher := launcher.NewClipboardLauncher(clipManager)
	appLauncher := launcher.NewAppsLauncher(appManager)
	cmdLauncher := launcher.NewShellLauncher(shellManager)

	settingsApp := settings.New(goclipDb)
	settingsApp.SetReloadAppsCallback(appLauncher.RedrawApps)

	log.Info("Starting listener")
	hotkeyListener := HotkeyListener(goclipDb, clipLauncher, appLauncher, cmdLauncher)
	hotkeyListener.Start()

	settingsApp.Run()
}
