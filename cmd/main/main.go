package main

import (
	"Goclip/clipboard"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/ui"
	"Goclip/ui/gtk"
	hook "github.com/robotn/gohook"
)

type GoclipListener struct {
	app ui.GoclipUI
	db  db.GoclipDB
}

func NewGoclipListener(goclipApp ui.GoclipUI, goclipDB db.GoclipDB) *GoclipListener {
	return &GoclipListener{app: goclipApp, db: goclipDB}
}

func (s *GoclipListener) StartHotkeyListener() {
	go s.startHotkeyListener()
}

func (s *GoclipListener) startHotkeyListener() {
	settings, err := s.db.GetSettings()
	if err != nil {
		settings = db.DefaultSettings()
	}
	hook.Register(hook.KeyDown, []string{settings.HookKey, settings.HookModKey}, func(event hook.Event) {
		s.app.ShowEntries()
	})
	start := hook.Start()
	<-hook.Process(start)
}

func main() {
	goclipDB := storm.New("/tmp/goclipdb")
	goclipCb := clipboard.New(goclipDB)
	goclipApp := gtk.New(goclipDB, goclipCb)
	goclipListener := NewGoclipListener(goclipApp, goclipDB)
	goclipCb.StartListener()
	goclipListener.StartHotkeyListener()
	goclipApp.Start()
}
