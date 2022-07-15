package cliputils

import (
	"Goclip/db"
	"Goclip/log"
	"Goclip/utils"
	"context"
	"github.com/go-vgo/robotgo"
	"golang.design/x/clipboard"
	"time"
)

type ClipboardManager struct {
	db       db.GoclipDB
	reloadCb func()
}

func NewClipboardManager(myDb db.GoclipDB) *ClipboardManager {
	return &ClipboardManager{db: myDb}
}

func (s *ClipboardManager) SetReloadHistoryCallback(f func()) {
	s.reloadCb = f
}

func (s *ClipboardManager) StartListener() {
	go s.startTextListener()
	go s.startImageListener()
}

func (s *ClipboardManager) startTextListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		log.Info("Got text: ", string(data))
		entry := &db.ClipboardEntry{
			Md5:       utils.Md5Digest(data),
			Mime:      "text/plain",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddClipboardEntry(entry)
		s.reloadCb()
	}
}

func (s *ClipboardManager) startImageListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	for data := range ch {
		log.Info("Got image: ", len(data))
		entry := &db.ClipboardEntry{
			Md5:       utils.Md5Digest(data),
			Mime:      "image/png",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddClipboardEntry(entry)
		s.reloadCb()
	}
}

func (s *ClipboardManager) WriteText(text string) {
	clipboard.Write(clipboard.FmtText, []byte(text))
}

func (s *ClipboardManager) WriteImage(data []byte) {
	clipboard.Write(clipboard.FmtImage, data)
}

func (s *ClipboardManager) WriteEntry(entry *db.ClipboardEntry) {
	if entry.IsText() {
		log.Info("Writing text: ", string(entry.Data))
		s.WriteText(string(entry.Data))
		go func() {
			time.Sleep(time.Millisecond * 200)
			robotgo.KeyDown("ctrl")
			robotgo.KeyDown("shift")
			robotgo.KeyDown("v")
			robotgo.KeyUp("v")
			robotgo.KeyUp("shift")
			robotgo.KeyUp("ctrl")
		}()
	} else if entry.IsImage() {
		clipboard.Write(clipboard.FmtImage, entry.Data)
	} else {
		log.Warning("Warning: Invalid entry mimetype: ", entry.Mime)
	}
}

func (s *ClipboardManager) GetEntries() []*db.ClipboardEntry {
	var newEntries []*db.ClipboardEntry
	entries := s.db.GetClipboardEntries()
	for i := range entries {
		if entries[i].Starred {
			newEntries = append([]*db.ClipboardEntry{entries[i]}, newEntries...)
		} else {
			newEntries = append(newEntries, entries[i])
		}
	}
	return newEntries
}

func (s *ClipboardManager) GetEntry(md5 string) (*db.ClipboardEntry, error) {
	return s.db.GetClipboardEntry(md5)
}

func (s *ClipboardManager) ToggleStar(md5 string) error {
	entry, err := s.db.GetClipboardEntry(md5)
	if err != nil {
		return err
	}
	entry.Starred = !entry.Starred
	return s.db.AddClipboardEntry(entry)
}

func (s *ClipboardManager) DeleteEntry(md5 string) error {
	return s.db.DeleteClipboardEntry(md5)
}
