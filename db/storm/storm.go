package storm

import (
	"Goclip/db"
	"github.com/asdine/storm/v3"
	"log"
)

type GoclipDBStorm struct {
	db       *storm.DB
	settings *db.Settings
}

func New(fileName string) *GoclipDBStorm {
	myDb, err := storm.Open(fileName)
	if err != nil {
		log.Fatal("Error opening database:", err.Error())
	}
	return &GoclipDBStorm{db: myDb}
}

func (s *GoclipDBStorm) cleanup() {
	var entry db.Entry
	tot, err := s.db.Count(&entry)
	if err != nil {
		log.Println("Error getting db count:", err.Error())
		return
	}
	if tot > s.settings.MaxEntries {
		n := tot - s.settings.MaxEntries
		log.Println("Deleting", n, "entries.")
		var toDel []*db.Entry
		if err := s.db.AllByIndex("Timestamp", &toDel, storm.Limit(n)); err != nil {
			log.Println("Error getting db entries:", err.Error())
			return
		}
		for _, entry := range toDel {
			// log.Println("Deleting:", entry.Data)
			if err := s.db.DeleteStruct(entry); err != nil {
				log.Println("Error deleting db entry:", err.Error())
			}
		}
		log.Println("Db cleanup complete.")
	}
}

func (s *GoclipDBStorm) AddEntry(entry *db.Entry) error {
	if s.settings == nil {
		if _, err := s.GetSettings(); err != nil {
			s.settings = db.DefaultSettings()
		}
	}
	if err := s.db.Save(entry); err != nil {
		log.Println("Error adding db entry:", err.Error())
		return err
	}
	s.cleanup()
	return nil
}

func (s *GoclipDBStorm) DeleteEntry(md5 string) error {
	if err := s.db.DeleteStruct(&db.Entry{Md5: md5}); err != nil {
		log.Println("Error deleting db entry:", err.Error())
		return err
	}
	log.Println("Db entry deleted:", md5)
	return nil
}

func (s *GoclipDBStorm) GetEntry(md5 string) (*db.Entry, error) {
	entry := db.Entry{}
	if err := s.db.One("Md5", md5, &entry); err != nil {
		log.Println("Error getting db entry:", err.Error())
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) GetEntries() []*db.Entry {
	var entries []*db.Entry
	if err := s.db.AllByIndex("Timestamp", &entries, storm.Reverse()); err != nil {
		log.Println("Error getting db entries:", err.Error())
	}
	return entries
}

func (s *GoclipDBStorm) SaveSettings(settings *db.Settings) error {
	if err := s.db.Set("settings", 0, settings); err != nil {
		log.Println("Error saving settings to db:", err.Error())
		return err
	}
	s.settings = settings
	return nil
}

func (s *GoclipDBStorm) GetSettings() (*db.Settings, error) {
	settings := db.Settings{}
	if err := s.db.Get("settings", 0, &settings); err != nil {
		log.Println("Error getting settings from db:", err.Error())
		return nil, err
	}
	s.settings = &settings
	return &settings, nil
}

func (s *GoclipDBStorm) Drop() {
	s.db.Drop(&db.Entry{})
	s.db.Drop("settings")
}
