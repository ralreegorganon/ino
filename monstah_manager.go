package ino

type MonstahManager struct {
	monstahs []*Monstah
	DB       *DB
}

func NewMonstahManager(db *DB) (*MonstahManager, error) {
	feeds, err := db.GetFeeds()
	if err != nil {
		return nil, err
	}

	monstahs := make([]*Monstah, len(feeds))
	for i, feed := range feeds {
		m := NewMonstah(db)
		go m.Decode(feed.RemoteAddress)
		monstahs[i] = m
	}

	mm := &MonstahManager{
		monstahs: monstahs,
		DB:       db,
	}

	return mm, nil
}

func (mm *MonstahManager) Shutdown() {
	for _, m := range mm.monstahs {
		m.Shutdown()
	}
}
