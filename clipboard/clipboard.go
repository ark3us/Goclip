package clipboard

import (
	"Goclip/common"
	"Goclip/db"
	"context"
	"golang.design/x/clipboard"
	"log"
	"time"
)

type GoclipBoard struct {
	db db.GoclipDB
}

func New(myDb db.GoclipDB) *GoclipBoard {
	return &GoclipBoard{db: myDb}
}

func (s *GoclipBoard) StartListener() {
	go s.startTextListener()
	go s.startImageListener()
}

func (s *GoclipBoard) startTextListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		log.Println("Got text:", string(data))
		entry := &db.Entry{
			Md5:       common.Md5Digest(data),
			Mime:      "text/plain",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddEntry(entry)
	}
}

func (s *GoclipBoard) startImageListener() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	for data := range ch {
		log.Println("Got image:", len(data))
		entry := &db.Entry{
			Md5:       common.Md5Digest(data),
			Mime:      "image/png",
			Data:      data,
			Timestamp: time.Now(),
		}
		s.db.AddEntry(entry)
	}
}

func (s *GoclipBoard) WriteText(text string) {
	clipboard.Write(clipboard.FmtText, []byte(text))
}

func (s *GoclipBoard) WriteImage(data []byte) {
	clipboard.Write(clipboard.FmtImage, data)
}

func (s *GoclipBoard) WriteEntry(entry *db.Entry) {
	if entry.IsText() {
		s.WriteText(string(entry.Data))
	} else if entry.IsImage() {
		clipboard.Write(clipboard.FmtImage, entry.Data)
	} else {
		log.Println("Warning: Invalid entry mimetype:", entry.Mime)
	}
}
