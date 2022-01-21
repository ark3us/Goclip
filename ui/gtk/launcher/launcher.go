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
	"os/exec"
	"strings"
	"unsafe"
)

const imgMaxSize = 250
const textMaxSize = 100

type LauncherType int8

const (
	LauncherTypeClipboard LauncherType = iota
	LauncherTypeApps
)

type Row struct {
	Box      *gtk.Box
	Id       string
	DataType string
}

func (s *Row) IsSearchable() bool {
	return strings.Contains(s.DataType, "text") || strings.Contains(s.DataType, "app")
}

func (s *Row) DbEntryContains(db db.GoclipDB, text string) bool {
	if !s.IsSearchable() {
		return false
	}
	if strings.Contains(s.DataType, "text") {
		entry, err := db.GetEntry(s.Id)
		if err != nil {
			return false
		}
		return strings.Contains(strings.ToLower(string(entry.Data)), strings.ToLower(text))
	} else if strings.Contains(s.DataType, "app") {
		return strings.Contains(s.Id, text)
	}
	return false
}

func ImageFromBytes(data []byte) *gtk.Image {
	loader, err := gdk.PixbufLoaderNew()
	if err != nil {
		log.Error("Error loading Pixbuf", err)
		return nil
	}
	pixbuf, err := loader.WriteAndReturnPixbuf(data)
	if err != nil {
		log.Error("Error writing Pixbuf", err)
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
			log.Error("Error scaling image: ", err)
			return nil
		}
	}
	image, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		log.Error("Error loading image: ", err)
		return nil
	}
	return image
}

type GoclipLauncherGtk struct {
	db    db.GoclipDB
	clip  *clipboard.GoclipBoard
	lType LauncherType

	app        *gtk.Application
	contentWin *gtk.ApplicationWindow
	rows       []*Row
	searchBox  *gtk.Entry
}

func NewClipboardLauncher(myDb db.GoclipDB, myClip *clipboard.GoclipBoard) *GoclipLauncherGtk {
	return &GoclipLauncherGtk{
		db:    myDb,
		clip:  myClip,
		lType: LauncherTypeClipboard,
	}
}

func NewAppsLauncher(myDb db.GoclipDB) *GoclipLauncherGtk {
	return &GoclipLauncherGtk{
		db:    myDb,
		lType: LauncherTypeApps,
	}
}

func (s *GoclipLauncherGtk) Run() {
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
		log.Error("Error getting text from ClipboardEntry: ", err)
		return
	}
	for _, row := range s.rows {
		if text == "" {
			row.Box.Show()
		} else if row.DbEntryContains(s.db, text) {
			row.Box.Show()
		} else {
			row.Box.Hide()
		}
	}
}

func (s *GoclipLauncherGtk) drawSearchBox(layout *gtk.Box) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error drawing ClipboardEntry: ", err)
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

func (s *GoclipLauncherGtk) drawEntry(layout *gtk.Box, entry *db.ClipboardEntry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err)
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
			if row.Id == md5 {
				row.Box.Destroy()
			}
		}
	})
	row.Add(delButton)

	layout.Add(row)
	s.rows = append(s.rows, &Row{
		Box:      row,
		Id:       entry.Md5,
		DataType: entry.Mime,
	})
}

func (s *GoclipLauncherGtk) drawApp(layout *gtk.Box, entry *db.AppEntry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err)
	}
	tsLabel, err := gtk.LabelNew(common.TimeToString(entry.AccessTime))
	row.Add(tsLabel)

	entryButton, err := gtk.ButtonNew()
	entryButton.SetLabel(entry.Cmd)
	entryButton.SetHExpand(true)
	entryButton.Connect("clicked", func() {
		cmd := exec.Command("nohup", entry.Cmd)
		err := cmd.Start()
		if err != nil {
			log.Error("Command error: ", err)
		}
	})
	row.Add(entryButton)

	layout.Add(row)
	s.rows = append(s.rows, &Row{
		Box:      row,
		Id:       entry.Cmd,
		DataType: "app",
	})
}

func (s *GoclipLauncherGtk) ShowEntries() {
	var err error
	s.contentWin, err = gtk.ApplicationWindowNew(s.app)
	if err != nil {
		log.Fatal("Error creating content Window: ", err)
	}

	topBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	s.drawSearchBox(topBox)
	topBox.SetVExpand(false)

	contentBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	s.rows = nil
	if s.lType == LauncherTypeClipboard {
		for _, entry := range s.db.GetEntries() {
			s.drawEntry(contentBox, entry)
		}
	} else {
		for _, entry := range s.db.GetApps() {
			s.drawApp(contentBox, entry)
		}
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
