package settings

import (
	"Goclip/db"
	"Goclip/goclip"
	"Goclip/goclip/log"
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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
	gtk.Init(nil)
	menu, err := gtk.MenuNew()
	if err != nil {
		log.Error("Error creating menu:", err)
		return
	}
	item, err := gtk.MenuItemNewWithLabel("Settings")
	item.Connect("activate", func() {
		s.ShowSettings()
	})
	menu.Add(item)

	item, err = gtk.MenuItemNewWithLabel("Reload apps")
	item.Connect("activate", func() {
		go s.db.RefreshApps()
	})
	menu.Add(item)

	item, err = gtk.MenuItemNewWithLabel("Quit")
	item.Connect("activate", func() {
		gtk.MainQuit()
	})
	menu.Add(item)

	indicator := appindicator.New(goclip.AppId, "icon", appindicator.CategoryApplicationStatus)
	indicator.SetIconThemePath(".")
	indicator.SetTitle(goclip.AppName)
	indicator.SetLabel(goclip.AppName, goclip.AppName)
	indicator.SetStatus(appindicator.StatusActive)
	indicator.SetMenu(menu)
	menu.ShowAll()
	gtk.Main()
}

func (s *GoclipSettingsGtk) ShowSettings() {
	glib.IdleAdd(s.showSettings)
}

func (s *GoclipSettingsGtk) showSettings() {
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
	s.settingsWin.SetTitle(goclip.AppName + ": Settings")
	layout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	message, err := gtk.LabelNew("")

	grid, _ := gtk.GridNew()
	grid.SetRowSpacing(10)
	grid.SetColumnSpacing(10)
	label, err := gtk.LabelNew("Maximum entries:")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 0, 1, 1)

	inputMaxEntries, err := gtk.EntryNew()
	inputMaxEntries.SetText(strconv.Itoa(settings.MaxEntries))
	inputMaxEntries.SetHExpand(true)
	grid.Attach(inputMaxEntries, 1, 0, 1, 1)

	label, err = gtk.LabelNew("Clipboard Mod key:")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 1, 1, 1)

	inputModKey, err := gtk.EntryNew()
	inputModKey.SetText(settings.ClipboardModKey)
	grid.Attach(inputModKey, 1, 1, 1, 1)

	label, err = gtk.LabelNew("Clipboard key:")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 2, 1, 1)

	inputKey, err := gtk.EntryNew()
	inputKey.SetText(settings.ClipboardKey)
	grid.Attach(inputKey, 1, 2, 1, 1)

	label, err = gtk.LabelNew("Applications Mod key:")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 3, 1, 1)

	inputAppModKey, err := gtk.EntryNew()
	inputAppModKey.SetText(settings.AppsModKey)
	grid.Attach(inputAppModKey, 1, 3, 1, 1)

	label, err = gtk.LabelNew("Applications key:")
	label.SetHAlign(gtk.ALIGN_END)
	grid.Attach(label, 0, 4, 1, 1)

	inputAppKey, err := gtk.EntryNew()
	inputAppKey.SetText(settings.AppsKey)
	grid.Attach(inputAppKey, 1, 4, 1, 1)

	layout.Add(grid)

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
		appModKey, err := inputModKey.GetText()
		appHookKey, err := inputKey.GetText()
		settings.MaxEntries = n
		if modKey != settings.ClipboardModKey || hookKey != settings.ClipboardKey || appModKey != settings.AppsModKey || appHookKey != settings.AppsKey {
			settings.ClipboardModKey = modKey
			settings.ClipboardKey = hookKey
			settings.AppsModKey = appModKey
			settings.AppsKey = appHookKey
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
		s.db.DropAll()
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