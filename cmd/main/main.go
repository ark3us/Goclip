package main

import (
	"Goclip/common"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/ui"
	"Goclip/ui/gtk"
	"context"
	hook "github.com/robotn/gohook"
	"golang.design/x/clipboard"
	"log"
	"time"
)

type GoclipListener struct {
	app ui.GoclipUI
	db  db.GoclipDB
}

func NewGoclipListener(goclipApp ui.GoclipUI, goclipDB db.GoclipDB) *GoclipListener {
	return &GoclipListener{app: goclipApp, db: goclipDB}
}

func (s *GoclipListener) StartClipboardListener() {
	go s.startTextListener()
	go s.startImageListener()
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

func (s *GoclipListener) startTextListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		log.Println("Got text:", string(data))
		entry := &db.Entry{
			Md5:       common.Md5Digest(data),
			Mime:      "text/plain",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddEntry(entry)
	}
}

func (s *GoclipListener) startImageListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	for data := range ch {
		log.Println("Got image:", len(data))
		entry := &db.Entry{
			Md5:       common.Md5Digest(data),
			Mime:      "image/png",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddEntry(entry)
	}
}

func main() {
	goclipDB := storm.New("/tmp/goclipdb")
	goclipApp := gtk.New(goclipDB)
	goclipListener := NewGoclipListener(goclipApp, goclipDB)
	goclipListener.StartClipboardListener()
	goclipListener.StartHotkeyListener()
	goclipApp.Start()
}
