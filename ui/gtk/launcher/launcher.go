package launcher

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkwindow.h>
// static GdkWindow *toGdkWindow(void *p) { return (GDK_WINDOW(p)); }
import "C"
import (
	"Goclip/clipboard"
	"Goclip/common"
	"Goclip/common/log"
	"Goclip/db"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"strings"
	"unsafe"
)

const imgMaxSize = 250
const textMaxSize = 100

type ContentType int8

type Row struct {
	Box  *gtk.Box
	Md5  string
	Mime string
}

func ImageFromBytes(data []byte) *gtk.Image {
	loader, err := gdk.PixbufLoaderNew()
	if err != nil {
		log.Error("Error loading Pixbuf", err.Error())
		return nil
	}
	pixbuf, err := loader.WriteAndReturnPixbuf(data)
	if err != nil {
		log.Error("Error writing Pixbuf", err.Error())
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
			log.Error("Error scaling image: ", err.Error())
			return nil
		}
	}
	image, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		log.Error("Error loading image: ", err.Error())
		return nil
	}
	return image
}

type GoclipLauncherGtk struct {
	settings *db.Settings
	db       db.GoclipDB
	clip     *clipboard.GoclipBoard

	app        *gtk.Application
	contentWin *gtk.ApplicationWindow
	rows       []*Row
	searchBox  *gtk.Entry
}

func New(myDb db.GoclipDB, myClip *clipboard.GoclipBoard) *GoclipLauncherGtk {
	settings, err := myDb.GetSettings()
	if err != nil {
		log.Warning("Warning: Cannot load settings, using default.")
		settings = db.DefaultSettings()
	}
	return &GoclipLauncherGtk{
		db:       myDb,
		clip:     myClip,
		settings: settings,
	}
}

func (s *GoclipLauncherGtk) Start() {
	var err error
	log.Info("Starting App")
	s.app, err = gtk.ApplicationNew(common.AppId, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Cannot create Application: ", err)
	}
	s.app.Connect("activate", s.ShowEntries)
	s.app.Run(nil)
	log.Info("App closed")
}

func (s *GoclipLauncherGtk) Quit() {
	s.app.Quit()
}

func (s *GoclipLauncherGtk) onSearching() {
	text, err := s.searchBox.GetText()
	if err != nil {
		log.Error("Error getting text from Entry: ", err.Error())
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

func (s *GoclipLauncherGtk) drawSearchBox(layout *gtk.Box) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error drawing Entry: ", err.Error())
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

func (s *GoclipLauncherGtk) drawEntry(layout *gtk.Box, entry *db.Entry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err.Error())
	}
	tsLabel, err := gtk.LabelNew(common.TimeToString(entry.Timestamp))
	row.Add(tsLabel)

	entryButton, err := gtk.ButtonNew()
	if entry.IsText() {
		var text string
		if len(entry.Data) > textMaxSize {
			text = string(append(entry.Data[:textMaxSize], " ..."...))
		} else {
			text = string(entry.Data)
		}
		entryButton.SetLabel(text)
	} else if entry.IsImage() {
		image := ImageFromBytes(entry.Data)
		if image != nil {
			entryButton.SetImage(image)
		}
	} else {
		log.Warning("Warning: invalid entry type:", entry.Mime)
		return
	}
	entryButton.SetHExpand(true)
	entryButton.Connect("clicked", func() {
		s.clip.WriteEntry(entry)
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
				row.Box.Destroy()
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

func (s *GoclipLauncherGtk) ShowEntries() {
	var err error
	s.contentWin, err = gtk.ApplicationWindowNew(s.app)
	if err != nil {
		log.Fatal("Error creating content Window: ", err.Error())
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

func (s *GoclipLauncherGtk) onFocusOut() {
	s.contentWin.Destroy()
}
