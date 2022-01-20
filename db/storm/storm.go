package storm

import (
	"Goclip/common/log"
	"Goclip/db"
	"github.com/asdine/storm/v3"
)

type GoclipDBStorm struct {
	dbFile string
}

func New(dbFile string) *GoclipDBStorm {
	return &GoclipDBStorm{dbFile: dbFile}
}

func (s *GoclipDBStorm) cleanup(myDb *storm.DB) error {
	settings, err := s.getSettings(myDb)
	if err != nil {
		settings = db.DefaultSettings()
	}

	var entry db.Entry
	tot, err := myDb.Count(&entry)
	if err != nil {
		log.Error("Error getting db count: ", err.Error())
		return err
	}
	if tot > settings.MaxEntries {
		n := tot - settings.MaxEntries
		log.Info("Deleting ", n, " entries.")
		var toDel []*db.Entry
		if err := myDb.AllByIndex("Timestamp", &toDel, storm.Limit(n)); err != nil {
			log.Error("Error getting db entries:", err.Error())
			return err
		}
		for _, entry := range toDel {
			// log.Println("Deleting:", entry.Data)
			if err := myDb.DeleteStruct(entry); err != nil {
				log.Error("Error deleting db entry: ", err.Error())
			}
		}
		log.Info("Db cleanup complete.")
	}
	return nil
}

func (s *GoclipDBStorm) AddEntry(entry *db.Entry) error {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err)
		return err
	}
	defer myDb.Close()

	if err := myDb.Save(entry); err != nil {
		log.Error("Error adding db entry: ", err.Error())
		return err
	}
	return s.cleanup(myDb)
}

func (s *GoclipDBStorm) DeleteEntry(md5 string) error {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return err
	}
	defer myDb.Close()

	if err := myDb.DeleteStruct(&db.Entry{Md5: md5}); err != nil {
		log.Error("Error deleting db entry: ", err.Error())
		return err
	}
	log.Info("Db entry deleted:", md5)
	return nil
}

func (s *GoclipDBStorm) GetEntry(md5 string) (*db.Entry, error) {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return nil, err
	}
	defer myDb.Close()

	entry := db.Entry{}
	if err := myDb.One("Md5", md5, &entry); err != nil {
		log.Error("Error getting db entry:", err.Error())
		return nil, err
	}
	return &entry, nil
}

func (s *GoclipDBStorm) GetEntries() []*db.Entry {
	var entries []*db.Entry
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return entries
	}
	defer myDb.Close()

	if err := myDb.AllByIndex("Timestamp", &entries, storm.Reverse()); err != nil {
		log.Error("Error getting db entries: ", err.Error())
	}
	return entries
}

func (s *GoclipDBStorm) SaveSettings(settings *db.Settings) error {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return err
	}
	defer myDb.Close()

	if err := myDb.Set("settings", 0, settings); err != nil {
		log.Error("Error saving settings to db: ", err.Error())
		return err
	}
	settings = settings
	return nil
}

func (s *GoclipDBStorm) GetSettings() (*db.Settings, error) {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return nil, err
	}
	defer myDb.Close()
	return s.getSettings(myDb)
}

func (s *GoclipDBStorm) getSettings(myDb *storm.DB) (*db.Settings, error) {
	settings := db.Settings{}
	if err := myDb.Get("settings", 0, &settings); err != nil {
		log.Error("Error getting settings from db: ", err.Error())
		return nil, err
	}
	return &settings, nil
}

func (s *GoclipDBStorm) Drop() error {
	myDb, err := storm.Open(s.dbFile)
	if err != nil {
		log.Error("Error opening database: ", err.Error())
		return err
	}
	defer myDb.Close()

	if err := myDb.Drop(&db.Entry{}); err != nil {
		log.Error("Error dropping database: ", err)
	}
	if err := myDb.Drop("settings"); err != nil {
		log.Error("Error dropping database: ", err)
	}
	return nil
}
