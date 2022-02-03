package launcher

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkwindow.h>
// static GdkWindow *toGdkWindow(void *p) { return (GDK_WINDOW(p)); }
import "C"
import (
	"Goclip/apputils"
	"Goclip/cliputils"
	"Goclip/db"
	"Goclip/log"
	"Goclip/shellutils"
	"Goclip/ui"
	"Goclip/utils"
	_ "embed"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"strings"
	"time"
	"unsafe"
)

const windowWidth = 500
const imgMaxSize = 250
const iconMaxSize = 25
const textMaxSize = 100

type LauncherType int8

const (
	LauncherTypeClipboard LauncherType = iota
	LauncherTypeApps
	LauncherTypeShell
)

type Row struct {
	Box      *gtk.Box
	Id       string
	MimeType string
	IsApp    bool
	IsClip   bool
	IsShell  bool
}

func (s *Row) IsSearchable() bool {
	return strings.Contains(s.MimeType, "text") || s.IsApp
}

func (s *Row) IsText() bool {
	return strings.Contains(s.MimeType, "text")
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
	lType        LauncherType
	title        string
	clipManager  *cliputils.ClipboardManager
	shellManager *shellutils.ShellManager
	appManager   *apputils.AppManager

	app        *gtk.Application
	contentWin *gtk.Window
	rows       []*Row
	searchBox  *gtk.Entry
	contentBox *gtk.Box
	cmdBox     *gtk.Box
}

func NewClipboardLauncher(myClip *cliputils.ClipboardManager) ui.GoclipLauncher {
	return &GoclipLauncherGtk{
		clipManager: myClip,
		lType:       LauncherTypeClipboard,
		title:       utils.AppName + ": Clipboard",
	}
}

func NewAppsLauncher(appManager *apputils.AppManager) ui.GoclipLauncher {
	return &GoclipLauncherGtk{
		appManager: appManager,
		lType:      LauncherTypeApps,
		title:      utils.AppName + ": Applications",
	}
}

func NewShellLauncher(shellManager *shellutils.ShellManager) ui.GoclipLauncher {
	return &GoclipLauncherGtk{
		lType:        LauncherTypeShell,
		title:        utils.AppName + ": Shell",
		shellManager: shellManager,
	}
}

func (s *GoclipLauncherGtk) Quit() {
	s.app.Quit()
}

func (s *GoclipLauncherGtk) handleCompletions(text string) {
	if s.cmdBox != nil {
		s.cmdBox.Destroy()
	}
	if newText, err := shellutils.ExpandUserDir(text); err == nil {
		s.searchBox.SetText(newText)
		s.searchBox.SetPosition(-1)
	}
	s.cmdBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	histBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	label, _ := gtk.LabelNew("Command history")
	label.SetSizeRequest(windowWidth/2, 0)
	histBox.Add(label)
	shellBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	label, _ = gtk.LabelNew("Shell commands")
	label.SetSizeRequest(windowWidth/2, 0)
	shellBox.Add(label)

	if text != "" {
		completions := s.shellManager.GetShellCompletions(text)
		// log.Info("Completions: ", len(completions))
		for _, compl := range completions {
			if compl.Cmd == "" {
				continue
			}
			button, _ := gtk.ButtonNew()
			button.SetHExpand(true)
			if len(compl.Cmd) > textMaxSize/2 {
				button.SetLabel(compl.Cmd[:textMaxSize/2] + " ...")
			} else {
				button.SetLabel(compl.Cmd)
			}
			cmd := compl.Cmd
			button.Connect("focus-in-event", func() {
				s.searchBox.SetText(cmd)
				s.searchBox.ShowAll()
			})
			button.Connect("clicked", func() {
				s.searchBox.GrabFocus()
			})
			button.Connect("button-press-event", func(btn *gtk.Button, evt *gdk.Event) {
				btnEvt := gdk.EventButton{Event: evt}
				if btnEvt.Type() == gdk.EVENT_BUTTON_PRESS {
					if btnEvt.Button() == gdk.BUTTON_SECONDARY {
						s.contentWin.Destroy()
						if cmd != "" {
							shellutils.Exec(cmd, true)
							return
						}
					}
				}
			})
			if compl.IsHistory {
				histBox.Add(button)
			} else {
				shellBox.Add(button)
			}
		}
	}
	s.cmdBox.Add(histBox)
	s.cmdBox.Add(shellBox)
	s.contentBox.Add(s.cmdBox)
	s.contentBox.ShowAll()
}

func (s *GoclipLauncherGtk) rowContains(row *Row, text string) bool {
	if !row.IsSearchable() {
		return false
	}
	if row.IsApp {
		return strings.Contains(strings.ToLower(row.Id), strings.ToLower(text))
	}
	if row.IsText() {
		if entry, err := s.clipManager.GetEntry(row.Id); err == nil {
			return strings.Contains(strings.ToLower(string(entry.Data)), strings.ToLower(text))
		}
	}
	return false
}

func (s *GoclipLauncherGtk) onSearching() {
	text, err := s.searchBox.GetText()
	if err != nil {
		log.Error("Error getting text from ClipboardEntry: ", err)
		return
	}
	switch s.lType {
	case LauncherTypeShell:
		s.handleCompletions(text)
	default:
		for _, row := range s.rows {
			if text == "" {
				row.Box.Show()
			} else if s.rowContains(row, text) {
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
	if s.lType == LauncherTypeShell {
		s.searchBox.Connect("activate", func() {
			cmd, _ := s.searchBox.GetText()
			s.contentWin.Destroy()
			if cmd != "" {
				shellutils.Exec(cmd, true)
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
	tsLabel, err := gtk.LabelNew(utils.TimeToString(entry.Timestamp, false))
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
				if entry, err := s.clipManager.GetEntry(md5); err == nil {
					s.clipManager.WriteEntry(entry)
				}
			} else if btnEvt.Button() == gdk.BUTTON_SECONDARY {
				log.Info("Right click")
				if entry, err := s.clipManager.GetEntry(md5); err == nil {
					shellutils.OpenEntry(entry)
				}
			}
			s.contentWin.Destroy()
		}
	})
	row.Add(entryButton)

	delButton, err := gtk.ButtonNew()
	delButton.SetLabel("X")
	delButton.Connect("clicked", func() {
		s.clipManager.DeleteEntry(md5)
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
		MimeType: entry.Mime,
		IsClip:   true,
	})
}

func (s *GoclipLauncherGtk) drawApp(entry *db.AppEntry) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Error creating box: ", err)
	}
	tsLabel, err := gtk.LabelNew(utils.TimeToString(entry.AccessTime, false))
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
		s.appManager.ExecEntry(entry)
	})
	row.Add(entryButton)

	s.contentBox.Add(row)
	s.rows = append(s.rows, &Row{
		Box:   row,
		Id:    entry.Exec,
		IsApp: true,
	})
}

func (s *GoclipLauncherGtk) ShowEntries() {
	glib.IdleAdd(s.showEntries)
}

func (s *GoclipLauncherGtk) drawEntries() {
	switch s.lType {
	case LauncherTypeClipboard:
		s.contentBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
		for _, entry := range s.clipManager.GetEntries() {
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
	s.appManager.LoadApps()
	s.contentBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	log.Info("Redrawing apps")
	for _, entry := range s.appManager.GetApps() {
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
	s.contentWin.SetDefaultSize(windowWidth, windowWidth)
	s.contentWin.SetSkipTaskbarHint(true)
	s.contentWin.SetTypeHint(gdk.WINDOW_TYPE_HINT_UTILITY)
	s.contentWin.SetKeepAbove(true)
	s.contentWin.SetPosition(gtk.WIN_POS_MOUSE)
	s.contentWin.Connect("focus-out-event", s.onFocusOut)
	s.contentWin.Connect("key-press-event", s.onKeyPress)
	s.contentWin.Connect("destroy", func() {
		switch s.lType {
		case LauncherTypeApps:
			go s.RedrawApps()
		case LauncherTypeShell:
			go s.shellManager.LoadHistory()
		}
	})

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
