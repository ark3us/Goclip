package main

import (
	"Goclip/clipboard"
	"Goclip/common/log"
	"Goclip/db"
	"Goclip/db/storm"
	"Goclip/ui"
	"Goclip/ui/gtk/launcher"
	"Goclip/ui/gtk/settings"
	hook "github.com/robotn/gohook"
	"os"
	"os/exec"
)

type GoclipListener struct {
	app      string
	db       db.GoclipDB
	cb       *clipboard.GoclipBoard
	settings ui.GoclipSettings
}

func NewGoclipListener(goclipApp string, goclipDB db.GoclipDB, goclipCb *clipboard.GoclipBoard) *GoclipListener {
	return &GoclipListener{
		app: goclipApp,
		db:  goclipDB,
		cb:  goclipCb,
	}
}

func (s *GoclipListener) Start() {
	go s.startHotkeyListener()
	go s.cb.StartListener()
}

func (s *GoclipListener) startHotkeyListener() {
	sets, err := s.db.GetSettings()
	if err != nil {
		sets = db.DefaultSettings()
	}
	hook.Register(hook.KeyDown, []string{sets.HookKey, sets.HookModKey}, func(event hook.Event) {
		go func() {
			cmd := exec.Command(s.app, "launcher")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			err := cmd.Run()
			if err != nil {
				log.Error("Launcher error: ", err)
			}
		}()
	})
	start := hook.Start()
	<-hook.Process(start)
}

func main() {
	// log.Debug = true
	goclipDB := storm.New("/tmp/goclipdb")
	goclipCb := clipboard.New(goclipDB)
	if len(os.Args) > 1 && os.Args[1] == "launcher" {
		log.Info("Launcher started")
		goclipApp := launcher.New(goclipDB, goclipCb)
		goclipApp.Start()
	} else {
		log.Info("Listener started")
		ex, err := os.Executable()
		if err != nil {
			log.Fatal("Error getting executable:", err)
		}
		goclipApp := settings.New(goclipDB)
		goclipListener := NewGoclipListener(ex, goclipDB, goclipCb)
		goclipListener.Start()
		goclipApp.Start()
	}
}
