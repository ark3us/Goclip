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
	Exec       string `storm:"id"`
	File       string
	Name       string
	Icon       string
	Terminal   bool
	AccessTime time.Time `storm:"index"`
}

type ShellEntry struct {
	Cmd       string `storm:"id"`
	IsHistory bool
	IsShell   bool
}

type Settings struct {
	MaxEntries      int
	ClipboardModKey string
	ClipboardKey    string
	AppsModKey      string
	AppsKey         string
	CmdModKey       string
	CmdKey          string
}

func DefaultSettings() *Settings {
	return &Settings{
		MaxEntries:      100,
		ClipboardModKey: "win",
		ClipboardKey:    "v",
		AppsModKey:      "win",
		AppsKey:         "c",
		CmdModKey:       "win",
		CmdKey:          "x",
	}
}

type GoclipDB interface {
	AddClipboardEntry(entry *ClipboardEntry) error
	DeleteClipboardEntry(md5 string) error
	GetClipboardEntry(md5 string) (*ClipboardEntry, error)
	GetClipboardEntries() []*ClipboardEntry

	AddAppEntries([]*AppEntry) error
	GetAppEntries() []*AppEntry
	GetAppEntry(cmd string) (*AppEntry, error)
	UpdateAppEntry(entry *AppEntry)

	AddShellEntries([]*ShellEntry) error
	GetShellEntries(cmd string, limit int) ([]*ShellEntry, error)

	GetSettings() (*Settings, error)
	SaveSettings(settings *Settings) error

	DropAll() error
	DropSettings() error
	DropClipboard() error
	DropApps() error
	DropShell() error
}
