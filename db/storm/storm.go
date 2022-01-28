package storm

import (
	"Goclip/db"
	"Goclip/log"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/protobuf"
	"github.com/asdine/storm/v3/q"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GoclipDBStorm struct {
	clipDb  *storm.DB
	appDb   *storm.DB
	shellDb *storm.DB
	setsDb  *storm.DB
}

func New(dbDir string) (*GoclipDBStorm, error) {
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Error("Error opening db directory: ", err)
		return nil, err
	}
	clipDb, err := openDb(filepath.Join(dbDir, "gcDb_clipboard"))
	if err != nil {
		return nil, err
	}
	appDb, err := openDb(filepath.Join(dbDir, "gcDb_apps"))
	if err != nil {
		return nil, err
	}
	shellDb, err := openDb(filepath.Join(dbDir, "gcDb_shell"))
	if err != nil {
		return nil, err
	}
	setsDb, err := openDb(filepath.Join(dbDir, "gcDb_settings"))
	if err != nil {
		return nil, err
	}
	return &GoclipDBStorm{
		clipDb:  clipDb,
		appDb:   appDb,
		shellDb: shellDb,
		setsDb:  setsDb,
	}, nil
}

func openDb(fn string) (*storm.DB, error) {
	myDb, err := storm.Open(fn, storm.Codec(protobuf.Codec))
	if err != nil {
		log.Error("Error opening database: ", fn, " - ", err)
		return nil, err
	}
	return myDb, nil
}

func (s *GoclipDBStorm) cleanup() error {
	settings, err := s.GetSettings()
	if err != nil {
		settings = db.DefaultSettings()
	}

	var entry db.ClipboardEntry
	tot, err := s.clipDb.Count(&entry)
	if err != nil {
		log.Error("Error getting db count: ", err)
		return err
	}
	if tot > settings.MaxEntries {
		n := tot - settings.MaxEntries
		log.Info("Deleting ", n, " entries.")
		var toDel []*db.ClipboardEntry
		if err := s.clipDb.AllByIndex("Timestamp", &toDel, storm.Limit(n)); err != nil {
			log.Error("Error getting db entries:", err)
			return err
		}
		for _, entry := range toDel {
			// log.Println("Deleting:", entry.Data)
			if err := s.clipDb.DeleteStruct(entry); err != nil {
				log.Error("Error deleting db entry: ", err)
			}
		}
		log.Info("Db cleanup complete.")
	}
	return nil
}

func (s *GoclipDBStorm) AddClipboardEntry(entry *db.ClipboardEntry) error {
	if err := s.clipDb.Save(entry); err != nil {
		log.Error("Error adding db entry: ", err)
		return err
	}
	return s.cleanup()
}

func (s *GoclipDBStorm) DeleteClipboardEntry(md5 string) error {
	if err := s.clipDb.DeleteStruct(&db.ClipboardEntry{Md5: md5}); err != nil {
		log.Error("Error deleting db entry: ", err)
		return err
	}
	log.Info("Db entry deleted:", md5)
	return nil
}

func (s *GoclipDBStorm) GetClipboardEntry(md5 string) (*db.ClipboardEntry, error) {
	entry := db.ClipboardEntry{}
	if err := s.clipDb.One("Md5", md5, &entry); err != nil {
		log.Error("Error getting db entry:", err)
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) GetClipboardEntries() []*db.ClipboardEntry {
	var entries []*db.ClipboardEntry
	if err := s.clipDb.AllByIndex("Timestamp", &entries, storm.Reverse()); err != nil {
		log.Error("Error getting db entries: ", err)
	}
	return entries
}

func (s *GoclipDBStorm) SaveSettings(settings *db.Settings) error {
	if err := s.setsDb.Set("settings", 0, settings); err != nil {
		log.Error("Error saving settings to db: ", err)
		return err
	}
	return nil
}

func (s *GoclipDBStorm) GetSettings() (*db.Settings, error) {
	settings := db.Settings{}
	if err := s.setsDb.Get("settings", 0, &settings); err != nil {
		log.Error("Error getting settings from db: ", err)
		return nil, err
	}
	return &settings, nil
}

func (s *GoclipDBStorm) DropSettings() error {
	log.Info("Dropping settings...")
	if err := s.setsDb.Drop("settings"); err != nil {
		log.Error("Error dropping settings: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) DropClipboard() error {
	log.Info("Dropping clipboard...")
	if err := s.clipDb.Drop(&db.ClipboardEntry{}); err != nil {
		log.Error("Error dropping clipboard: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) DropApps() error {
	log.Info("Dropping apps...")
	if err := s.appDb.Drop(&db.AppEntry{}); err != nil {
		log.Error("Error dropping apps: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) DropShell() error {
	log.Info("Dropping shell history...")
	if err := s.shellDb.Drop(&db.ShellEntry{}); err != nil {
		log.Error("Error dropping shell history: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) DropAll() error {
	log.Info("Dropping everything...")
	if err := s.DropClipboard(); err != nil {
		return err
	}
	if err := s.DropApps(); err != nil {
		return err
	}
	if err := s.DropShell(); err != nil {
		return err
	}
	if err := s.DropSettings(); err != nil {
		return err
	}
	return nil
}

func appEntriesContain(entries []*db.AppEntry, entry *db.AppEntry) bool {
	for i := range entries {
		if entries[i].Exec == entry.Exec {
			return true
		}
	}
	return false
}

func (s *GoclipDBStorm) AddAppEntries(newEntries []*db.AppEntry) error {
	log.Info("Removing old apps...")
	tx, err := s.appDb.Begin(true)
	if err != nil {
		log.Error("Cannot start transaction: ", err)
		return err
	}
	removed := 0
	var oldEntries []*db.AppEntry
	if err := tx.All(&oldEntries); err != nil {
		log.Warning("Cannot get old entries: ", err)
	}
	var toExclude []*db.AppEntry
	for i := range oldEntries {
		if !appEntriesContain(newEntries, oldEntries[i]) {
			if err := tx.DeleteStruct(oldEntries[i]); err != nil {
				log.Warning("Cannot delete old entry: ", err)
			} else {
				removed++
			}
		} else {
			toExclude = append(toExclude, oldEntries[i])
		}
	}
	log.Info("Old apps removed: ", removed)

	log.Info("Adding new apps...")
	added := 0
	for i := range newEntries {
		if !appEntriesContain(toExclude, newEntries[i]) {
			log.Info("New:", newEntries[i].Exec)
			if err := tx.Save(newEntries[i]); err != nil {
				log.Error("Cannot save entry, aborting: ", err)
				tx.Rollback()
				return err
			}
			added++
		}
	}
	log.Info("Refresh complete, added apps: ", added)
	if err := tx.Commit(); err != nil {
		log.Error("Cannot commit transaction: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) AddShellEntries(entries []*db.ShellEntry) error {
	// Bulk insert without transaction is SLOW because it waits i/o for each Save
	tx, err := s.shellDb.Begin(true)
	if err != nil {
		log.Error("Error starting transaction: ", err)
		return err
	}
	if err = tx.Drop(&db.ShellEntry{}); err != nil {
		log.Warning("Error dropping shell history: ", err)
	}
	for i := range entries {
		if err = tx.Save(entries[i]); err != nil {
			log.Error("Cannot save entry, aborting: ", err)
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Error("Cannot commit transaction: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) GetShellEntries(cmd string, limit int) ([]*db.ShellEntry, error) {
	query := s.shellDb.Select(q.Re("Cmd", "(?i).*"+cmd+".*")).Limit(limit)
	var results []*db.ShellEntry
	if err := query.Find(&results); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		log.Error("Error finding completions: ", err)
		return nil, err
	}
	return results, nil
}

func (s *GoclipDBStorm) GetAppEntries() []*db.AppEntry {
	log.Info("Getting all apps...")
	var entries []*db.AppEntry
	if err := s.appDb.AllByIndex("AccessTime", &entries, storm.Reverse()); err != nil {
		log.Error("Error getting db entries: ", err)
	}
	log.Info("Got apps.")
	return entries
}

func (s *GoclipDBStorm) GetAppEntry(cmd string) (*db.AppEntry, error) {
	entry := db.AppEntry{}
	if err := s.appDb.One("Exec", cmd, &entry); err != nil {
		log.Error("Error getting db entry:", err)
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) UpdateAppEntry(entry *db.AppEntry) {
	entry.AccessTime = time.Now()
	if err := s.appDb.Update(entry); err != nil {
		log.Warning("Error updating entry: ", err)
	}
}
