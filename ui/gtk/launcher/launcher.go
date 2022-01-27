package launcher

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkwindow.h>
// static GdkWindow *toGdkWindow(void *p) { return (GDK_WINDOW(p)); }
import "C"
import (
	"Goclip/clipboard"
	"Goclip/db"
	"Goclip/goclip"
	"Goclip/goclip/cmdutils"
	"Goclip/goclip/log"
	_ "embed"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"os/user"
	"strings"
	"time"
	"unsafe"
)

const imgMaxSize = 250
const iconMaxSize = 25
const textMaxSize = 100

type LauncherType int8

const (
	LauncherTypeClipboard LauncherType = iota
	LauncherTypeApps
	LauncherTypeCmd
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

func ImageFromBytes(data []byte, maxSize int) *gtk.Image {
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
	if pixbuf.GetHeight() > maxSize || pixbuf.GetWidth() > maxSize {
		var newWidth, newHeight = 0, 0
		if pixbuf.GetHeight() == pixbuf.GetWidth() {
			newWidth, newHeight = maxSize, maxSize
		} else if pixbuf.GetHeight() > pixbuf.GetWidth() {
			newHeight = maxSize
			newWidth = maxSize * pixbuf.GetWidth() / pixbuf.GetHeight()
		} else {
			newWidth = maxSize
			newHeight = maxSize * pixbuf.GetHeight() / pixbuf.GetWidth()
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

func ImageFromFile(fn string, maxSize int) *gtk.Image {
	if fn == "" {
		return nil
	}
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Error("Error opening icon: ", fn)
		return nil
	}
	return ImageFromBytes(data, maxSize)
}

type GoclipLauncherGtk struct {
	db    db.GoclipDB
	clip  *clipboard.GoclipBoard
	lType LauncherType
	title string

	app        *gtk.Application
	contentWin *gtk.Window
	rows       []*Row
	searchBox  *gtk.Entry
	contentBox *gtk.Box
	cmdBox     *gtk.Box
}

func NewClipboardLauncher(myDb db.GoclipDB, myClip *clipboard.GoclipBoard) *GoclipLauncherGtk {
	return &GoclipLauncherGtk{
		db:    myDb,
		clip:  myClip,
		lType: LauncherTypeClipboard,
		title: goclip.AppName + ": Clipboard",
	}
}

func NewAppsLauncher(myDb db.GoclipDB) *GoclipLauncherGtk {
	return &GoclipLauncherGtk{
		db:    myDb,
		lType: LauncherTypeApps,
		title: goclip.AppName + ": Applications",
	}
}

func NewCmdLauncher() *GoclipLauncherGtk {
	return &GoclipLauncherGtk{
		lType: LauncherTypeCmd,
		title: goclip.AppName + ": Shell",
	}
}

func (s *GoclipLauncherGtk) Quit() {
	s.app.Quit()
}

func (s *GoclipLauncherGtk) handleCompletions(text string) {
	if s.cmdBox != nil {
		s.cmdBox.Destroy()
	}
	if strings.Contains(text, "~/") {
		usr, err := user.Current()
		if err == nil {
			text = strings.Replace(text, "~/", usr.HomeDir+"/", -1)
			s.searchBox.SetText(text)
			s.searchBox.SetPosition(-1)
		}
	}
	s.cmdBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if text != "" {
		completions := cmdutils.GetCompletions(text)
		// log.Info("Completions: ", len(completions))
		for _, compl := range completions {
			if compl == "" {
				continue
			}
			label, _ := gtk.ButtonNew()
			label.SetHExpand(true)
			label.SetLabel(compl)
			_compl := compl
			label.Connect("focus-in-event", func() {
				s.searchBox.SetText(_compl)
				s.searchBox.ShowAll()
			})
			label.Connect("clicked", func() {
				s.searchBox.GrabFocus()
			})
			s.cmdBox.Add(label)
		}
	}
	s.contentBox.Add(s.cmdBox)
	s.contentBox.ShowAll()
}

func (s *GoclipLauncherGtk) onSearching() {
	text, err := s.searchBox.GetText()
	if err != nil {
		log.Error("Error getting text from ClipboardEntry: ", err)
		return
	}
	switch s.lType {
	case LauncherTypeCmd:
		s.handleCompletions(text)
	default:
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
	s.searchBox.Connect("grab-focus", func() {
		go func() {
			time.Sleep(time.Millisecond)
			s.searchBox.SelectRegion(0, 0)
			s.searchBox.SetPosition(-1)
		}()
	})
	if s.lType == LauncherTypeCmd {
		s.searchBox.Connect("activate", func() {
			cmd, _ := s.searchBox.GetText()
			s.contentWin.Destroy()
			if cmd != "" {
				cmdutils.Exec(cmd, true)
			}
		})
	}
	s.searchBox.GrabFocus()

	row.Add(s.searchBox)
	layout.Add(row)
}

func (s *GoclipLauncherGtk) drawEntry(entry *db.ClipboardEntry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err)
	}
	tsLabel, err := gtk.LabelNew(goclip.TimeToString(entry.Timestamp, false))
	row.Add(tsLabel)

	entryButton, err := gtk.ButtonNew()
	entryButton.SetHExpand(true)
	if entry.IsText() {
		var text string
		if len(entry.Data) > textMaxSize {
			text = string(append(entry.Data[:textMaxSize], " ..."...))
		} else {
			text = string(entry.Data)
		}
		entryButton.SetLabel(text)
	} else if entry.IsImage() {
		image := ImageFromBytes(entry.Data, imgMaxSize)
		if image != nil {
			entryButton.SetImage(image)
		}
	} else {
		log.Warning("Warning: invalid entry type:", entry.Mime)
		return
	}

	md5 := entry.Md5
	entryButton.Connect("button-press-event", func(btn *gtk.Button, evt *gdk.Event) {
		btnEvt := gdk.EventButton{Event: evt}
		if btnEvt.Type() == gdk.EVENT_BUTTON_PRESS {
			if btnEvt.Button() == gdk.BUTTON_PRIMARY {
				log.Info("Left click")
				if entry, err := s.db.GetEntry(md5); err == nil {
					s.clip.WriteEntry(entry)
				}
			} else if btnEvt.Button() == gdk.BUTTON_SECONDARY {
				log.Info("Right click")
				if entry, err := s.db.GetEntry(md5); err == nil {
					cmdutils.ExecEntry(entry)
				}
			}
			s.contentWin.Destroy()
		}
	})
	row.Add(entryButton)

	delButton, err := gtk.ButtonNew()
	delButton.SetLabel("X")
	delButton.Connect("clicked", func() {
		s.db.DeleteEntry(md5)
		for _, row := range s.rows {
			if row.Id == md5 {
				row.Box.Destroy()
			}
		}
	})
	row.Add(delButton)

	s.contentBox.Add(row)
	s.rows = append(s.rows, &Row{
		Box:      row,
		Id:       entry.Md5,
		DataType: entry.Mime,
	})
}

func (s *GoclipLauncherGtk) drawApp(entry *db.AppEntry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err)
	}
	tsLabel, err := gtk.LabelNew(goclip.TimeToString(entry.AccessTime, false))
	row.Add(tsLabel)

	image := ImageFromFile(entry.Icon, iconMaxSize)
	img, _ := gtk.ImageNew()
	if image != nil {
		img.SetFromPixbuf(image.GetPixbuf())
	}
	img.SetSizeRequest(iconMaxSize, iconMaxSize)
	row.Add(img)

	entryButton, err := gtk.ButtonNew()
	entryButton.SetLabel(entry.Name)
	entryButton.SetHExpand(true)
	entryButton.Connect("clicked", func() {
		s.contentWin.Destroy()
		log.Info("Entry: ", entry.File, " Exec: ", entry.Exec)
		cmdutils.Exec(entry.Exec, entry.Terminal)
		s.db.UpdateAppAccess(entry)
	})
	row.Add(entryButton)

	s.contentBox.Add(row)
	s.rows = append(s.rows, &Row{
		Box:      row,
		Id:       entry.Exec,
		DataType: "app",
	})
}

func (s *GoclipLauncherGtk) ShowEntries() {
	glib.IdleAdd(s.showEntries)
}

func (s *GoclipLauncherGtk) drawEntries() {
	switch s.lType {
	case LauncherTypeClipboard:
		s.contentBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
		for _, entry := range s.db.GetEntries() {
			s.drawEntry(entry)
		}
	case LauncherTypeApps:
		if s.contentBox == nil {
			s.RedrawApps()
		}
	default:
		s.contentBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	}
}

func (s *GoclipLauncherGtk) RedrawApps() {
	s.contentBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	log.Info("Redrawing apps")
	for _, entry := range s.db.GetApps() {
		s.drawApp(entry)
	}
}

func (s *GoclipLauncherGtk) showEntries() {
	var err error
	if s.contentWin != nil {
		s.contentWin.Destroy()
	}
	s.contentWin, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Error creating content Window: ", err)
	}

	topBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	s.drawSearchBox(topBox)
	topBox.SetVExpand(false)

	s.drawEntries()
	contentScroll, err := gtk.ScrolledWindowNew(nil, nil)
	contentScroll.Add(s.contentBox)
	contentScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	contentScroll.SetVExpand(true)

	contentLayout, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	contentLayout.Add(topBox)
	contentLayout.Add(contentScroll)

	s.contentWin.Add(contentLayout)
	s.contentWin.SetTitle(s.title)
	s.contentWin.SetDefaultSize(500, 500)
	s.contentWin.SetSkipTaskbarHint(true)
	s.contentWin.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
	s.contentWin.SetKeepAbove(true)
	s.contentWin.SetPosition(gtk.WIN_POS_MOUSE)
	s.contentWin.Connect("focus-out-event", s.onFocusOut)
	s.contentWin.Connect("key-press-event", s.onKeyPress)

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

func (s *GoclipLauncherGtk) onKeyPress(widget *gtk.Window, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)
	if keyEvent.KeyVal() == gdk.KEY_Escape {
		s.contentWin.Destroy()
	}
}
