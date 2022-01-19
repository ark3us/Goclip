package gtk

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkwindow.h>
// static GdkWindow *toGdkWindow(void *p) { return (GDK_WINDOW(p)); }
import "C"
import (
	"Goclip/common"
	"Goclip/db"
	"github.com/getlantern/systray"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"golang.design/x/clipboard"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

const imgMaxSize = 250
const textMaxSize = 100

type ContentType int8

const (
	ContentClipboard ContentType = iota
	ContentApps
)

type Row struct {
	Box  *gtk.Box
	Md5  string
	Mime string
}

func ImageFromBytes(data []byte) *gtk.Image {
	loader, err := gdk.PixbufLoaderNew()
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	pixbuf, err := loader.WriteAndReturnPixbuf(data)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	if pixbuf.GetHeight() > imgMaxSize || pixbuf.GetWidth() > imgMaxSize {
		var newWidth, newHeight = 0, 0
		if pixbuf.GetHeight() == pixbuf.GetWidth() {
			newWidth, newHeight = imgMaxSize, imgMaxSize
		} else if pixbuf.GetHeight() > pixbuf.GetWidth() {
			newHeight = imgMaxSize
			newWidth = imgMaxSize * pixbuf.GetWidth() / pixbuf.GetHeight()
		} else {
			newWidth = imgMaxSize
			newHeight = imgMaxSize * pixbuf.GetHeight() / pixbuf.GetWidth()
		}
		pixbuf, err = pixbuf.ScaleSimple(newWidth, newHeight, gdk.INTERP_HYPER)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
	}
	image, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return image
}

type GoclipUIGtk struct {
	settings *db.Settings
	db       db.GoclipDB

	app         *gtk.Application
	contentWin  *gtk.ApplicationWindow
	settingsWin *gtk.ApplicationWindow
	rows        []*Row
	searchBox   *gtk.Entry
}

func New(myDb db.GoclipDB) *GoclipUIGtk {
	settings, err := myDb.GetSettings()
	if err != nil {
		log.Println("Warning: Cannot load settings, using default.")
		settings = db.DefaultSettings()
	}
	return &GoclipUIGtk{
		db:       myDb,
		settings: settings,
	}
}

func (s *GoclipUIGtk) Start() {
	log.Println("Starting App")
	var err error
	s.app, err = gtk.ApplicationNew(common.AppId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.app.Connect("activate", s.startSystray)
	s.app.Run(os.Args)
	log.Println("App closed")
}

func (s *GoclipUIGtk) Quit() {
	gtk.MainQuit()
}

func (s *GoclipUIGtk) startSystray() {
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
			s.Quit()
		}()
	}, func() {
	})
}

func (s *GoclipUIGtk) onSearching() {
	text, err := s.searchBox.GetText()
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, row := range s.rows {
		if text == "" {
			row.Box.Show()
		} else if !strings.Contains(row.Mime, "text") {
			row.Box.Hide()
		} else {
			entry, err := s.db.GetEntry(row.Md5)
			if err != nil {
				continue
			}
			entryText := strings.ToLower(string(entry.Data))
			if !strings.Contains(entryText, strings.ToLower(text)) {
				row.Box.Hide()
			} else {
				row.Box.Show()
			}
		}
	}
}

func (s *GoclipUIGtk) drawSearchBox(layout *gtk.Box) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal(err.Error())
	}
	label, err := gtk.LabelNew("Search:")
	row.Add(label)

	s.searchBox, err = gtk.EntryNew()
	s.searchBox.SetHExpand(true)
	s.searchBox.Connect("key-release-event", s.onSearching)
	s.searchBox.GrabFocus()

	row.Add(s.searchBox)
	layout.Add(row)
}

func (s *GoclipUIGtk) drawEntry(layout *gtk.Box, entry *db.Entry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal(err.Error())
	}
	tsLabel, err := gtk.LabelNew(common.TimeToString(entry.Timestamp))
	row.Add(tsLabel)

	entryButton, err := gtk.ButtonNew()
	fmt := clipboard.FmtText
	if strings.Contains(entry.Mime, "text") {
		var text string
		if len(entry.Data) > textMaxSize {
			text = string(append(entry.Data[:textMaxSize], " ..."...))
		} else {
			text = string(entry.Data)
		}
		entryButton.SetLabel(text)
	} else {
		image := ImageFromBytes(entry.Data)
		if image != nil {
			entryButton.SetImage(image)
		}
		fmt = clipboard.FmtImage
	}
	entryButton.SetHExpand(true)
	data := entry.Data
	entryButton.Connect("clicked", func() {
		clipboard.Write(fmt, data)
		s.contentWin.Close()
	})
	row.Add(entryButton)

	delButton, err := gtk.ButtonNew()
	delButton.SetLabel("X")
	md5 := entry.Md5
	delButton.Connect("clicked", func() {
		s.db.DeleteEntry(md5)
		for _, row := range s.rows {
			if row.Md5 == md5 {
				row.Box.Hide()
			}
		}
	})
	row.Add(delButton)

	layout.Add(row)
	s.rows = append(s.rows, &Row{
		Box:  row,
		Md5:  entry.Md5,
		Mime: entry.Mime,
	})
}

func (s *GoclipUIGtk) ShowEntries() {
	var err error
	if s.contentWin != nil {
		// TODO: Destroy() can crash the application
		s.contentWin.Hide()
	}
	s.contentWin, err = gtk.ApplicationWindowNew(s.app)
	if err != nil {
		log.Fatal(err.Error())
	}

	topBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	s.drawSearchBox(topBox)
	topBox.SetVExpand(false)

	contentBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	s.rows = nil
	for _, entry := range s.db.GetEntries() {
		s.drawEntry(contentBox, entry)
	}

	contentScroll, err := gtk.ScrolledWindowNew(nil, nil)
	contentScroll.Add(contentBox)
	contentScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	contentScroll.SetVExpand(true)

	contentLayout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	contentLayout.Add(topBox)
	contentLayout.Add(contentScroll)

	s.contentWin.Add(contentLayout)
	s.contentWin.SetTitle(common.AppName)
	s.contentWin.SetDefaultSize(500, 500)
	s.contentWin.SetSkipTaskbarHint(true)
	s.contentWin.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
	s.contentWin.SetKeepAbove(true)
	s.contentWin.SetPosition(gtk.WIN_POS_MOUSE)
	s.contentWin.Connect("focus-out-event", s.onFocusOut)

	// Trick needed to grab the focus
	s.contentWin.PresentWithTime(gdk.CURRENT_TIME)
	w, _ := s.contentWin.GetWindow()
	p := unsafe.Pointer(w.GObject)
	C.gdk_window_focus(C.toGdkWindow(p), gdk.CURRENT_TIME)

	s.searchBox.GrabFocus()
	s.contentWin.ShowAll()
}

func (s *GoclipUIGtk) onFocusOut() bool {
	// TODO: Destroy() can crash the application
	s.contentWin.Hide()
	return true
}

func (s *GoclipUIGtk) ShowSettings() {
	var err error
	if s.settingsWin != nil {
		s.settingsWin.Destroy()
	}
	s.settingsWin, err = gtk.ApplicationWindowNew(s.app)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.settingsWin.SetTitle(common.AppName + ": Settings")
	layout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	message, err := gtk.LabelNew("")

	label, err := gtk.LabelNew("Maximum entries:")
	layout.Add(label)

	inputMaxEntries, err := gtk.EntryNew()
	inputMaxEntries.SetText(strconv.Itoa(s.settings.MaxEntries))
	layout.Add(inputMaxEntries)

	label, err = gtk.LabelNew("Shortcut - mod key:")
	layout.Add(label)
	inputModKey, err := gtk.EntryNew()
	inputModKey.SetText(s.settings.HookModKey)
	layout.Add(inputModKey)

	label, err = gtk.LabelNew("Shortcut - key:")
	layout.Add(label)
	inputKey, err := gtk.EntryNew()
	inputKey.SetText(s.settings.HookKey)
	layout.Add(inputKey)

	save, err := gtk.ButtonNew()
	save.SetLabel("Save")
	save.Connect("clicked", func() {
		maxEntries, err := inputMaxEntries.GetText()
		n, err := strconv.Atoi(maxEntries)
		if err != nil {
			log.Println(err.Error())
			return
		}
		modKey, err := inputModKey.GetText()
		hookKey, err := inputKey.GetText()
		s.settings.MaxEntries = n
		if modKey != s.settings.HookModKey || hookKey != s.settings.HookKey {
			s.settings.HookModKey = modKey
			s.settings.HookKey = hookKey
			message.SetLabel("Application restart required")
		}
		s.db.SaveSettings(s.settings)
	})
	layout.Add(save)

	resetSettings, err := gtk.ButtonNew()
	resetSettings.SetLabel("Reset settings")
	resetSettings.Connect("clicked", func() {
		s.settings = db.DefaultSettings()
		s.db.SaveSettings(s.settings)
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
