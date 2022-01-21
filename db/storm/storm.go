package storm

import (
	"Goclip/db"
	"Goclip/goclip/apputils"
	"Goclip/goclip/log"
	"github.com/asdine/storm/v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GoclipDBStorm struct {
	dbFile string
	db     *storm.DB
}

func New(dbFile string) *GoclipDBStorm {
	return &GoclipDBStorm{dbFile: dbFile}
}

func (s *GoclipDBStorm) openDb() error {
	var err error
	s.db, err = storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err)
		return err
	}
	return nil
}

func (s *GoclipDBStorm) closeDb() {
	if err := s.db.Close(); err != nil {
		log.Error("Error closing database: ", err)
	}
}

func (s *GoclipDBStorm) cleanup() error {
	settings, err := s.getSettings()
	if err != nil {
		settings = db.DefaultSettings()
	}

	var entry db.ClipboardEntry
	tot, err := s.db.Count(&entry)
	if err != nil {
		log.Error("Error getting db count: ", err)
		return err
	}
	if tot > settings.MaxEntries {
		n := tot - settings.MaxEntries
		log.Info("Deleting ", n, " entries.")
		var toDel []*db.ClipboardEntry
		if err := s.db.AllByIndex("Timestamp", &toDel, storm.Limit(n)); err != nil {
			log.Error("Error getting db entries:", err)
			return err
		}
		for _, entry := range toDel {
			// log.Println("Deleting:", entry.Data)
			if err := s.db.DeleteStruct(entry); err != nil {
				log.Error("Error deleting db entry: ", err)
			}
		}
		log.Info("Db cleanup complete.")
	}
	return nil
}

func (s *GoclipDBStorm) AddEntry(entry *db.ClipboardEntry) error {
	if err := s.openDb(); err != nil {
		return err
	}
	defer s.closeDb()

	if err := s.db.Save(entry); err != nil {
		log.Error("Error adding db entry: ", err)
		return err
	}
	return s.cleanup()
}

func (s *GoclipDBStorm) DeleteEntry(md5 string) error {
	if err := s.openDb(); err != nil {
		return err
	}
	defer s.closeDb()

	if err := s.db.DeleteStruct(&db.ClipboardEntry{Md5: md5}); err != nil {
		log.Error("Error deleting db entry: ", err)
		return err
	}
	log.Info("Db entry deleted:", md5)
	return nil
}

func (s *GoclipDBStorm) GetEntry(md5 string) (*db.ClipboardEntry, error) {
	if err := s.openDb(); err != nil {
		return nil, err
	}
	defer s.closeDb()

	entry := db.ClipboardEntry{}
	if err := s.db.One("Md5", md5, &entry); err != nil {
		log.Error("Error getting db entry:", err)
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) GetEntries() []*db.ClipboardEntry {
	if err := s.openDb(); err != nil {
		return nil
	}
	defer s.closeDb()

	var entries []*db.ClipboardEntry
	if err := s.db.AllByIndex("Timestamp", &entries, storm.Reverse()); err != nil {
		log.Error("Error getting db entries: ", err)
	}
	return entries
}

func (s *GoclipDBStorm) SaveSettings(settings *db.Settings) error {
	if err := s.openDb(); err != nil {
		return err
	}
	defer s.closeDb()

	if err := s.db.Set("settings", 0, settings); err != nil {
		log.Error("Error saving settings to db: ", err)
		return err
	}
	return nil
}

func (s *GoclipDBStorm) GetSettings() (*db.Settings, error) {
	if err := s.openDb(); err != nil {
		return nil, err
	}
	defer s.closeDb()
	return s.getSettings()
}

func (s *GoclipDBStorm) getSettings() (*db.Settings, error) {
	settings := db.Settings{}
	if err := s.db.Get("settings", 0, &settings); err != nil {
		log.Error("Error getting settings from db: ", err)
		return nil, err
	}
	return &settings, nil
}

func (s *GoclipDBStorm) Drop() error {
	if err := s.openDb(); err != nil {
		return err
	}
	defer s.closeDb()

	if err := s.db.Drop(&db.ClipboardEntry{}); err != nil {
		log.Error("Error dropping database: ", err)
	}
	if err := s.db.Drop("settings"); err != nil {
		log.Error("Error dropping database: ", err)
	}
	return nil
}

func (s *GoclipDBStorm) RefreshApps() error {
	log.Info("Refreshing apps...")
	if err := s.openDb(); err != nil {
		return err
	}
	defer s.closeDb()

	pathEnv := os.Getenv("XDG_DATA_DIRS")
	paths := strings.Split(pathEnv, ":")
	for _, path := range paths {
		path = filepath.Join(path, "applications")
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Warning("Cannot read dir content: ", err)
			continue
		}
		n := 0
		for _, finfo := range files {
			ffile := filepath.Join(path, finfo.Name())
			entries, err := apputils.ParseDesktopFile(ffile)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if err := s.db.Save(entry); err != nil {
					log.Error("Cannot save entry, aborting: ", err)
					return err
				}
			}
			n++
		}
		log.Info(path, ": ", n)
	}
	tot, _ := s.db.Count(&db.AppEntry{})
	log.Info("Refresh complete, added apps: ", tot)
	return nil
}

func (s *GoclipDBStorm) GetApps() []*db.AppEntry {
	if err := s.openDb(); err != nil {
		return nil
	}
	defer s.closeDb()
	log.Info("Getting all apps...")
	var entries []*db.AppEntry
	if err := s.db.AllByIndex("AccessTime", &entries, storm.Reverse()); err != nil {
		log.Error("Error getting db entries: ", err)
	}
	log.Info("Got apps.")
	return entries
}

func (s *GoclipDBStorm) GetApp(cmd string) (*db.AppEntry, error) {
	if err := s.openDb(); err != nil {
		return nil, err
	}
	defer s.closeDb()

	entry := db.AppEntry{}
	if err := s.db.One("Exec", cmd, &entry); err != nil {
		log.Error("Error getting db entry:", err)
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) UpdateAppAccess(entry *db.AppEntry) {
	if err := s.openDb(); err != nil {
		return
	}
	defer s.closeDb()
	entry.AccessTime = time.Now()
	if err := s.db.Update(entry); err != nil {
		log.Warning("Error updating entry: ", err)
	}
}
