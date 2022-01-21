package settings

import (
	"Goclip/common"
	"Goclip/common/log"
	"Goclip/db"
	"github.com/getlantern/systray"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"strconv"
)

type GoclipSettingsGtk struct {
	db          db.GoclipDB
	settingsWin *gtk.Window
}

func New(goclipDB db.GoclipDB) *GoclipSettingsGtk {
	return &GoclipSettingsGtk{db: goclipDB}
}

func (s *GoclipSettingsGtk) Run() {
	systray.Run(func() {
		data, _ := ioutil.ReadFile("icon.png")
		systray.SetIcon(data)
		systray.SetTitle(common.AppName)
		mSettings := systray.AddMenuItem("Settings", "Application settings")
		go func() {
			for {
				<-mSettings.ClickedCh
				s.ShowSettings()
			}
		}()
		mQuit := systray.AddMenuItem("Quit", "Quit app")
		go func() {
			<-mQuit.ClickedCh
			systray.Quit()
		}()
	}, func() {
	})
}

func (s *GoclipSettingsGtk) ShowSettings() {
	var err error
	if s.settingsWin != nil {
		s.settingsWin.Destroy()
	}
	settings, err := s.db.GetSettings()
	if err != nil {
		settings = db.DefaultSettings()
	}

	s.settingsWin, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Error creating settings Window: ", err.Error())
	}
	s.settingsWin.SetTitle(common.AppName + ": Settings")
	layout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	message, err := gtk.LabelNew("")

	label, err := gtk.LabelNew("Maximum entries:")
	layout.Add(label)

	inputMaxEntries, err := gtk.EntryNew()
	inputMaxEntries.SetText(strconv.Itoa(settings.MaxEntries))
	layout.Add(inputMaxEntries)

	label, err = gtk.LabelNew("Shortcut - mod key:")
	layout.Add(label)
	inputModKey, err := gtk.EntryNew()
	inputModKey.SetText(settings.ClipboardModKey)
	layout.Add(inputModKey)

	label, err = gtk.LabelNew("Shortcut - key:")
	layout.Add(label)
	inputKey, err := gtk.EntryNew()
	inputKey.SetText(settings.ClipboardKey)
	layout.Add(inputKey)

	save, err := gtk.ButtonNew()
	save.SetLabel("Save")
	save.Connect("clicked", func() {
		maxEntries, err := inputMaxEntries.GetText()
		n, err := strconv.Atoi(maxEntries)
		if err != nil {
			log.Error("Invalid ClipboardEntry value: ", err.Error())
			message.SetText("Invalid value")
			return
		}
		modKey, err := inputModKey.GetText()
		hookKey, err := inputKey.GetText()
		settings.MaxEntries = n
		if modKey != settings.ClipboardModKey || hookKey != settings.ClipboardKey {
			settings.ClipboardModKey = modKey
			settings.ClipboardKey = hookKey
			message.SetLabel("Application restart required")
		}
		s.db.SaveSettings(settings)
	})
	layout.Add(save)

	resetSettings, err := gtk.ButtonNew()
	resetSettings.SetLabel("Reset settings")
	resetSettings.Connect("clicked", func() {
		settings = db.DefaultSettings()
		s.db.SaveSettings(settings)
		message.SetLabel("Application restart required")
	})
	layout.Add(resetSettings)

	resetDb, err := gtk.ButtonNew()
	resetDb.SetLabel("Reset Database")
	resetDb.Connect("clicked", func() {
		s.db.Drop()
		message.SetLabel("Application restart required")
	})
	layout.Add(resetDb)

	layout.Add(message)
	s.settingsWin.Add(layout)
	s.settingsWin.SetDefaultSize(500, 500)
	s.settingsWin.SetPosition(gtk.WIN_POS_MOUSE)
	s.settingsWin.SetKeepAbove(true)
	s.settingsWin.SetSkipTaskbarHint(true)
	s.settingsWin.ShowAll()
}
