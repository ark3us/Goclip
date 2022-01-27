package cliputils

import (
	"Goclip/db"
	"Goclip/goclip"
	"Goclip/goclip/log"
	"context"
	"golang.design/x/clipboard"
	"time"
)

type ClipboardManager struct {
	db db.GoclipDB
}

func NewClipboardManager(myDb db.GoclipDB) *ClipboardManager {
	return &ClipboardManager{db: myDb}
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
			Md5:       goclip.Md5Digest(data),
			Mime:      "text/plain",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddClipboardEntry(entry)
	}
}

func (s *ClipboardManager) startImageListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	for data := range ch {
		log.Info("Got image: ", len(data))
		entry := &db.ClipboardEntry{
			Md5:       goclip.Md5Digest(data),
			Mime:      "image/png",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddClipboardEntry(entry)
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
		s.WriteText(string(entry.Data))
	} else if entry.IsImage() {
		clipboard.Write(clipboard.FmtImage, entry.Data)
	} else {
		log.Warning("Warning: Invalid entry mimetype: ", entry.Mime)
	}
}

func (s *ClipboardManager) GetEntries() []*db.ClipboardEntry {
	return s.db.GetClipboardEntries()
}

func (s *ClipboardManager) GetEntry(md5 string) (*db.ClipboardEntry, error) {
	return s.db.GetClipboardEntry(md5)
}

func (s *ClipboardManager) DeleteEntry(md5 string) error {
	return s.db.DeleteClipboardEntry(md5)
}
