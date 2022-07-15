package settings

import (
	"Goclip/db"
	"Goclip/log"
	"Goclip/ui"
	"Goclip/utils"
	_ "embed"
	"fyne.io/systray"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"strconv"
	"strings"
)

//go:embed Goclip.png
var iconData []byte

const (
	relIconDir  = ".local/share/icons/hicolor/512x512/apps"
	relIconFile = "Goclip.png"
	iconName    = "Goclip"
)

type GoclipSettingsGtk struct {
	db                db.GoclipDB
	settingsWin       *gtk.Window
	mainGrid          *gtk.Grid
	message           *gtk.Label
	gridRows          int
	reloadAppsCb      func()
	currSettings      *db.Settings
	inputMaxEntries   *gtk.Entry
	inputClipHookKey  *gtk.Entry
	inputAppHookKey   *gtk.Entry
	inputShellHookKey *gtk.Entry

	clipLauncher ui.GoclipLauncher
	appLauncher  ui.GoclipLauncher
	cmdLauncher  ui.GoclipLauncher
}

func New(goclipDB db.GoclipDB, clipLauncher, appLauncher, shellLauncher ui.GoclipLauncher) ui.GoclipSettings {
	return &GoclipSettingsGtk{
		db: goclipDB, clipLauncher: clipLauncher, appLauncher: appLauncher, cmdLauncher: shellLauncher}
}

func (s *GoclipSettingsGtk) SetReloadAppsCallback(callback func()) {
	s.reloadAppsCb = callback
}

func (s *GoclipSettingsGtk) Run() {
	gtk.Init(nil)
	go systray.Run(s.onReady, onExit)
	gtk.Main()
}

func (s *GoclipSettingsGtk) onReady() {
	systray.SetTitle("GoClip")
	systray.SetIcon(iconData)
	mClip := systray.AddMenuItem("Clipboard", "")
	mApp := systray.AddMenuItem("Apps", "")
	mShell := systray.AddMenuItem("Shell", "")
	mSettings := systray.AddMenuItem("Settings", "")
	mReload := systray.AddMenuItem("Reload Apps", "")
	mQuit := systray.AddMenuItem("Quit", "")
	if s.reloadAppsCb != nil {
		log.Info("Reloading apps...")
		go s.reloadAppsCb()
	}
	for {
		select {
		case <-mClip.ClickedCh:
			s.clipLauncher.ShowEntries()
		case <-mApp.ClickedCh:
			s.appLauncher.ShowEntries()
		case <-mShell.ClickedCh:
			s.cmdLauncher.ShowEntries()
		case <-mSettings.ClickedCh:
			s.ShowSettings()
		case <-mReload.ClickedCh:
			if s.reloadAppsCb != nil {
				go s.reloadAppsCb()
			} else {
				log.Error("No callback set")
			}
		case <-mQuit.ClickedCh:
			gtk.MainQuit()
		}
	}
}

func onExit() {
}

func (s *GoclipSettingsGtk) ShowSettings() {
	glib.IdleAdd(s.showSettings)
}

func (s *GoclipSettingsGtk) drawClipboardSettings() {
	label, _ := gtk.LabelNew("Clipboard launcher settings")
	s.mainGrid.Attach(label, 0, s.gridRows, 2, 1)
	s.gridRows++

	label, _ = gtk.LabelNew("Maximum entries:")
	label.SetHAlign(gtk.ALIGN_END)
	s.mainGrid.Attach(label, 0, s.gridRows, 1, 1)

	s.inputMaxEntries, _ = gtk.EntryNew()
	s.inputMaxEntries.SetText(strconv.Itoa(s.currSettings.MaxEntries))
	s.inputMaxEntries.SetHExpand(true)
	s.mainGrid.Attach(s.inputMaxEntries, 1, s.gridRows, 1, 1)
	s.gridRows++

	label, _ = gtk.LabelNew("Shortcut:")
	label.SetHAlign(gtk.ALIGN_END)
	s.mainGrid.Attach(label, 0, s.gridRows, 1, 1)

	s.inputClipHookKey, _ = gtk.EntryNew()
	s.inputClipHookKey.SetText(s.currSettings.ClipboardShortcut)
	s.mainGrid.Attach(s.inputClipHookKey, 1, s.gridRows, 1, 1)
	s.gridRows++
}

func (s *GoclipSettingsGtk) drawAppSettings() {
	label, _ := gtk.LabelNew("App launcher settings")
	s.mainGrid.Attach(label, 0, s.gridRows, 2, 1)
	s.gridRows++

	label, _ = gtk.LabelNew("Shortcut:")
	label.SetHAlign(gtk.ALIGN_END)
	s.mainGrid.Attach(label, 0, s.gridRows, 1, 1)

	s.inputAppHookKey, _ = gtk.EntryNew()
	s.inputAppHookKey.SetText(s.currSettings.AppsShortcut)
	s.mainGrid.Attach(s.inputAppHookKey, 1, s.gridRows, 1, 1)
	s.gridRows++
}

func (s *GoclipSettingsGtk) drawShellSettings() {
	label, _ := gtk.LabelNew("Shell launcher settings")
	s.mainGrid.Attach(label, 0, s.gridRows, 2, 1)
	s.gridRows++

	label, _ = gtk.LabelNew("Shortcut:")
	label.SetHAlign(gtk.ALIGN_END)
	s.mainGrid.Attach(label, 0, s.gridRows, 1, 1)

	s.inputShellHookKey, _ = gtk.EntryNew()
	s.inputShellHookKey.SetText(s.currSettings.ShellShortcut)
	s.mainGrid.Attach(s.inputShellHookKey, 1, s.gridRows, 1, 1)
	s.gridRows++
}

func (s *GoclipSettingsGtk) parseShortcut(shortcut string) (string, string) {
	parts := strings.Split(shortcut, "+")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		s.showMessage("Invalid shortcut: " + shortcut)
		return "", ""
	}
	return parts[0], parts[1]
}

func (s *GoclipSettingsGtk) checkKeyHooks() {
	clipboardShortcut, _ := s.inputClipHookKey.GetText()
	appsShortcut, _ := s.inputAppHookKey.GetText()
	shellShortcut, _ := s.inputShellHookKey.GetText()

	if clipboardShortcut != s.currSettings.ClipboardShortcut || appsShortcut != s.currSettings.AppsShortcut || shellShortcut != s.currSettings.ShellShortcut {
		s.showMessage("Application restart required")
	}

	s.currSettings.ClipboardShortcut = clipboardShortcut
	s.currSettings.AppsShortcut = appsShortcut
	s.currSettings.ShellShortcut = shellShortcut
}

func (s *GoclipSettingsGtk) showMessage(text string) {
	s.message.SetMarkup("<span foreground=\"red\">" + text + "</span>")
}

func (s *GoclipSettingsGtk) showSettings() {
	var err error
	if s.settingsWin != nil {
		s.settingsWin.Destroy()
	}
	s.currSettings, err = s.db.GetSettings()
	if err != nil {
		s.currSettings = db.DefaultSettings()
	}

	s.settingsWin, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Error creating settings Window: ", err.Error())
	}
	s.settingsWin.SetTitle(utils.AppName + ": Settings")
	s.settingsWin.SetBorderWidth(10)
	mainLayout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	s.mainGrid, _ = gtk.GridNew()
	s.mainGrid.SetRowSpacing(10)
	s.mainGrid.SetColumnSpacing(10)
	s.gridRows = 0

	s.message, err = gtk.LabelNew("")

	empty, err := gtk.LabelNew("")
	mainLayout.Add(empty)

	s.drawClipboardSettings()
	s.drawAppSettings()
	s.drawShellSettings()

	mainLayout.Add(s.mainGrid)

	save, err := gtk.ButtonNew()
	save.SetLabel("Save")
	save.Connect("clicked", func() {
		maxEntriesStr, _ := s.inputMaxEntries.GetText()
		maxEntries, err := strconv.Atoi(maxEntriesStr)
		if err != nil {
			s.showMessage("Invalid value for Maximum clipboard entries")
			maxEntries = s.currSettings.MaxEntries
		}
		s.currSettings.MaxEntries = maxEntries
		s.checkKeyHooks()
		s.db.SaveSettings(s.currSettings)
	})
	mainLayout.Add(save)

	resetSettings, err := gtk.ButtonNew()
	resetSettings.SetLabel("Reset settings")
	resetSettings.Connect("clicked", func() {
		s.currSettings = db.DefaultSettings()
		s.db.SaveSettings(s.currSettings)
		s.showMessage("Application restart required")
	})
	mainLayout.Add(resetSettings)

	resetClip, err := gtk.ButtonNew()
	resetClip.SetLabel("Reset clipboard")
	resetClip.Connect("clicked", func() {
		s.currSettings = db.DefaultSettings()
		s.db.DropApps()
		s.showMessage("Application restart required")
	})
	mainLayout.Add(resetClip)

	resetDb, err := gtk.ButtonNew()
	resetDb.SetLabel("Reset entire database")
	resetDb.Connect("clicked", func() {
		s.db.DropAll()
		s.showMessage("Application restart required")
	})
	mainLayout.Add(resetDb)

	mainLayout.Add(s.message)
	s.settingsWin.Add(mainLayout)
	s.settingsWin.SetDefaultSize(500, 250)
	s.settingsWin.SetPosition(gtk.WIN_POS_MOUSE)
	s.settingsWin.SetKeepAbove(true)
	s.settingsWin.SetSkipTaskbarHint(true)
	s.settingsWin.ShowAll()
}
