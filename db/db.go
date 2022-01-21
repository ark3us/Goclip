package db

import (
	"strings"
	"time"
)

type ClipboardEntry struct {
	Md5       string    `storm:"id"`
	Timestamp time.Time `storm:"index"`
	Mime      string
	Data      []byte
}

func (s *ClipboardEntry) IsText() bool {
	return strings.Contains(s.Mime, "text")
}

func (s *ClipboardEntry) IsImage() bool {
	return strings.Contains(s.Mime, "image")
}

type AppEntry struct {
	Cmd        string    `storm:"id"`
	AccessTime time.Time `storm:"index"`
}

type Settings struct {
	MaxEntries      int
	ClipboardModKey string
	ClipboardKey    string
	AppsModKey      string
	AppsKey         string
}

func DefaultSettings() *Settings {
	return &Settings{
		MaxEntries:      100,
		ClipboardModKey: "alt",
		ClipboardKey:    "z",
		AppsModKey:      "alt",
		AppsKey:         "x",
	}
}

type GoclipDB interface {
	AddEntry(entry *ClipboardEntry) error
	DeleteEntry(md5 string) error
	GetEntry(md5 string) (*ClipboardEntry, error)
	GetEntries() []*ClipboardEntry

	RefreshApps() error
	GetApps() []*AppEntry
	GetApp(cmd string) (*AppEntry, error)

	GetSettings() (*Settings, error)
	SaveSettings(settings *Settings) error

	Drop() error
}
