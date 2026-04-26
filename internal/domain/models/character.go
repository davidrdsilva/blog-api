package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Character represents a recurring person in the Whitenest serial fiction.
// Characters are reused across multiple chapters via a many-to-many relation
// on the posts table. Skills follow a fixed canonical schema (see CharacterSkills).
type Character struct {
	ID          string          `gorm:"type:uuid;primaryKey" json:"id"`
	FullName    string          `gorm:"type:varchar(120);not null" json:"full_name"`
	ShortName   string          `gorm:"type:varchar(60);not null" json:"short_name"`
	Description string          `gorm:"type:text;not null" json:"description"`
	Occupation  string          `gorm:"type:varchar(120);not null" json:"occupation"`
	Location    string          `gorm:"type:varchar(160);not null" json:"location"`
	Portrait    string          `gorm:"type:varchar(2048);not null" json:"portrait"`
	Skills      CharacterSkills `gorm:"type:jsonb;not null" json:"skills"`
	CreatedAt   time.Time       `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (Character) TableName() string {
	return "characters"
}

func (c *Character) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// CharacterSkills is the fixed canonical skill set for every character. Values
// are stored as JSONB on the characters table and are required to be present
// for every key (the radar chart on the frontend assumes a fixed axis set).
type CharacterSkills struct {
	Melee      int `json:"melee"`
	Guns       int `json:"guns"`
	Stealth    int `json:"stealth"`
	Persuasion int `json:"persuasion"`
	Intellect  int `json:"intellect"`
	Endurance  int `json:"endurance"`
}

// Value implements driver.Valuer for JSONB persistence.
func (s CharacterSkills) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for JSONB retrieval.
func (s *CharacterSkills) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal CharacterSkills: invalid type")
	}
	return json.Unmarshal(bytes, s)
}

// Validate enforces 0..100 bounds on every skill axis.
func (s CharacterSkills) Validate() error {
	checks := []struct {
		name  string
		value int
	}{
		{"melee", s.Melee},
		{"guns", s.Guns},
		{"stealth", s.Stealth},
		{"persuasion", s.Persuasion},
		{"intellect", s.Intellect},
		{"endurance", s.Endurance},
	}
	for _, c := range checks {
		if c.value < 0 || c.value > 100 {
			return fmt.Errorf("skill %q must be between 0 and 100, got %d", c.name, c.value)
		}
	}
	return nil
}

// CharacterFilters holds filtering options for querying characters.
type CharacterFilters struct {
	Search string
}

// PostsCharacter is the explicit join row between posts and characters. The
// `position` column preserves the order in which characters were assigned to
// a chapter, so the frontend can render the cast in the author's intended
// order without resorting to alphabetical sorting.
type PostsCharacter struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	PostID      string `gorm:"type:uuid;not null;index" json:"post_id"`
	CharacterID string `gorm:"type:uuid;not null;index" json:"character_id"`
	Position    int    `gorm:"not null;default:0" json:"position"`
}

// TableName specifies the table name for GORM
func (PostsCharacter) TableName() string {
	return "posts_characters"
}

// BeforeCreate generates a UUID for new join rows.
func (pc *PostsCharacter) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == "" {
		pc.ID = uuid.New().String()
	}
	return nil
}
