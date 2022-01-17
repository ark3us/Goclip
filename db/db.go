package db

import "time"

type Entry struct {
	Md5       string    `storm:"id"`
	Timestamp time.Time `storm:"index"`
	Mime      string
	Data      []byte
}

type Settings struct {
	MaxEntries int
	HookModKey string
	HookKey    string
}

func DefaultSettings() *Settings {
	return &Settings{
		MaxEntries: 100,
		HookModKey: "alt",
		HookKey:    "z",
	}
}

type GoclipDB interface {
	AddEntry(entry *Entry) error
	DeleteEntry(md5 string) error
	GetEntry(md5 string) (*Entry, error)
	GetEntries() []*Entry
	GetSettings() (*Settings, error)
	SaveSettings(settings *Settings) error
	Drop()
}
