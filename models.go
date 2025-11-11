package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Key represents a display key that can be managed dynamically
type Key struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"not null" json:"name"`                 // User-defined name like "KEY0", "节目一"
	Position        int       `gorm:"not null;uniqueIndex" json:"position"` // Display order
	KeyType         string    `gorm:"not null" json:"key_type"`             // "program" or "lyrics"
	Status          string    `gorm:"default:CLOSED" json:"status"`         // CLOSED, OPENING, OPENED, CLOSING, PLAYING_FORWARD
	TransitionTime  float64   `gorm:"default:1.0" json:"transition_time"`   // Default transition time in seconds
	CurrentPresetID *uint     `json:"current_preset_id"`                    // Currently selected preset (program or song)
	Version         int       `gorm:"default:0" json:"version"`             // For optimistic locking
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Program represents program information (for "program" type keys)
type Program struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	KeyID     uint      `gorm:"not null;index" json:"key_id"`
	Num       string    `json:"num"`                      // Program number
	Name      string    `json:"name"`                     // Program name
	Person    string    `json:"person"`                   // Performer name
	Position  int       `gorm:"not null" json:"position"` // Order within the key
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Key Key `gorm:"foreignKey:KeyID;constraint:OnDelete:CASCADE" json:"-"`
}

// Song represents a song (for "lyrics" type keys)
type Song struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	KeyID          uint      `gorm:"not null;index" json:"key_id"`
	Name           string    `gorm:"not null" json:"name"`     // Song name
	CurrentLyricID *uint     `json:"current_lyric_id"`         // Currently displayed lyric line
	Position       int       `gorm:"not null" json:"position"` // Order within the key
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Key    Key     `gorm:"foreignKey:KeyID;constraint:OnDelete:CASCADE" json:"-"`
	Lyrics []Lyric `gorm:"foreignKey:SongID" json:"lyrics,omitempty"`
}

// Lyric represents a single line of lyrics
type Lyric struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SongID         uint      `gorm:"not null;index" json:"song_id"`
	Text           string    `gorm:"type:text" json:"text"`
	TransitionTime float64   `gorm:"default:1.0" json:"transition_time"` // Animation time for this line
	Position       int       `gorm:"not null" json:"position"`           // Order within the song
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Song Song `gorm:"foreignKey:SongID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides
func (Key) TableName() string {
	return "keys"
}

func (Program) TableName() string {
	return "programs"
}

func (Song) TableName() string {
	return "songs"
}

func (Lyric) TableName() string {
	return "lyrics"
}

// Database initialization
func initDB() error {
	var err error
	db, err = gorm.Open(sqlite.Open("live-titler.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto migrate all tables
	err = db.AutoMigrate(&Key{}, &Program{}, &Song{}, &Lyric{})
	if err != nil {
		return err
	}

	// Initialize with default data if tables are empty
	var count int64
	db.Model(&Key{}).Count(&count)
	if count == 0 {
		return initDefaultData()
	}

	return nil
}

// Initialize with sample data
func initDefaultData() error {
	// Create default keys
	keys := []Key{
		{
			Name:           "节目一",
			Position:       0,
			KeyType:        "program",
			Status:         "CLOSED",
			TransitionTime: 1.0,
		},
		{
			Name:           "节目二",
			Position:       1,
			KeyType:        "program",
			Status:         "CLOSED",
			TransitionTime: 1.0,
		},
		{
			Name:           "歌词一",
			Position:       2,
			KeyType:        "lyrics",
			Status:         "CLOSED",
			TransitionTime: 1.0,
		},
		{
			Name:           "歌词二",
			Position:       3,
			KeyType:        "lyrics",
			Status:         "CLOSED",
			TransitionTime: 1.0,
		},
	}

	// Create keys and related data in transaction
	return db.Transaction(func(tx *gorm.DB) error {
		for i := range keys {
			if err := tx.Create(&keys[i]).Error; err != nil {
				return err
			}

			if keys[i].KeyType == "program" {
				// Create sample programs
				programs := []Program{
					{
						KeyID:    keys[i].ID,
						Num:      "1",
						Name:     "示例节目",
						Person:   "表演者",
						Position: 0,
					},
					{
						KeyID:    keys[i].ID,
						Num:      "2",
						Name:     "第二个节目",
						Person:   "其他表演者",
						Position: 1,
					},
				}
				for j := range programs {
					if err := tx.Create(&programs[j]).Error; err != nil {
						return err
					}
				}
				// Set first program as current
				keys[i].CurrentPresetID = &programs[0].ID
				if err := tx.Save(&keys[i]).Error; err != nil {
					return err
				}
			} else if keys[i].KeyType == "lyrics" {
				// Create sample song with lyrics
				song := Song{
					KeyID:    keys[i].ID,
					Name:     "示例歌曲",
					Position: 0,
				}
				if err := tx.Create(&song).Error; err != nil {
					return err
				}

				lyrics := []Lyric{
					{
						SongID:         song.ID,
						Text:           "第一句歌词",
						TransitionTime: 1.0,
						Position:       0,
					},
					{
						SongID:         song.ID,
						Text:           "第二句歌词",
						TransitionTime: 1.0,
						Position:       1,
					},
					{
						SongID:         song.ID,
						Text:           "第三句歌词",
						TransitionTime: 1.0,
						Position:       2,
					},
				}
				for j := range lyrics {
					if err := tx.Create(&lyrics[j]).Error; err != nil {
						return err
					}
				}

				// Set first lyric as current and first song as current preset
				song.CurrentLyricID = &lyrics[0].ID
				if err := tx.Save(&song).Error; err != nil {
					return err
				}
				keys[i].CurrentPresetID = &song.ID
				if err := tx.Save(&keys[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
