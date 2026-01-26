package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// EditorJsContent represents the content structure from Editor.js
type EditorJsContent struct {
	Blocks  []EditorJsBlock `json:"blocks"`
	Time    int64           `json:"time"`    // Unix timestamp in milliseconds
	Version string          `json:"version"` // Editor.js version
}

// EditorJsBlock represents a single content block in Editor.js
type EditorJsBlock struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// Value implements the driver.Valuer interface for database storage
// Converts EditorJsContent to JSON for storage in JSONB column
func (c EditorJsContent) Value() (driver.Value, error) {
	if c.Blocks == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
// Converts JSON from database to EditorJsContent struct
func (c *EditorJsContent) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal EditorJsContent: invalid type")
	}

	return json.Unmarshal(bytes, c)
}
