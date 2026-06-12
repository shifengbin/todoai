package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// TerminalHistoryRecord is a persisted snapshot of terminal metadata and
// recent output history. It is saved to terminal-history.json alongside
// the existing project config and restored on application startup.
type TerminalHistoryRecord struct {
	TerminalID     string `json:"terminalId"`
	ProjectID      string `json:"projectId"`
	TodoID         string `json:"todoId,omitempty"`
	TodoProjectID  string `json:"todoProjectId,omitempty"`
	ShellName      string `json:"shellName"`
	State          string `json:"state"`
	CreatedAt      string `json:"createdAt"`
	LastSelectedAt string `json:"lastSelectedAt"`
	Output         string `json:"output,omitempty"`
}

// TerminalHistoryStore manages persistence of terminal history records
// to a JSON file on disk. It provides load, save, and atomic write
// operations and handles missing-file gracefully (empty history).
type TerminalHistoryStore struct {
	path string
}

// TerminalHistoryFile is the top-level structure serialized to disk.
type TerminalHistoryFile struct {
	Version int                     `json:"version"`
	Records []TerminalHistoryRecord `json:"records"`
}

// MaxTerminalOutputBytes is the hard per-terminal output history size
// limit. Output beyond this limit is trimmed from the beginning,
// keeping the most recent bytes.
const MaxTerminalOutputBytes = 200 * 1024 // 200 KB

// NewTerminalHistoryStore creates a store that reads from and writes to
// the given file path.
func NewTerminalHistoryStore(configDir string) *TerminalHistoryStore {
	return &TerminalHistoryStore{
		path: filepath.Join(configDir, "terminal-history.json"),
	}
}

// Load reads the terminal history file from disk. A missing file is
// treated as an empty history without error.
func (store *TerminalHistoryStore) Load() (TerminalHistoryFile, error) {
	data, err := os.ReadFile(store.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return TerminalHistoryFile{Version: 1}, nil
		}
		return TerminalHistoryFile{}, err
	}

	var history TerminalHistoryFile
	if err := json.Unmarshal(data, &history); err != nil {
		return TerminalHistoryFile{Version: 1}, nil
	}
	if history.Version == 0 {
		history.Version = 1
	}
	return history, nil
}

// Save atomically writes the terminal history file to disk using a
// temporary file and rename to avoid corruption on crash.
func (store *TerminalHistoryStore) Save(history TerminalHistoryFile) error {
	if history.Version == 0 {
		history.Version = 1
	}
	if history.Records == nil {
		history.Records = []TerminalHistoryRecord{}
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := store.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, store.path)
}

// AppendOutput appends data to a terminal's output history and trims
// the result to MaxTerminalOutputBytes, keeping only the most recent
// bytes. Returns the updated output string.
func AppendTerminalOutput(existing string, data string) string {
	combined := existing + data
	if len(combined) <= MaxTerminalOutputBytes {
		return combined
	}
	// Keep the most recent bytes by trimming from the beginning.
	// Walk forward to the first valid UTF-8 start byte after the cut point.
	cut := len(combined) - MaxTerminalOutputBytes
	for cut < len(combined) {
		if combined[cut] < 0x80 || combined[cut]&0xC0 == 0xC0 {
			break
		}
		cut++
	}
	if cut >= len(combined) {
		cut = len(combined) - MaxTerminalOutputBytes
	}
	return combined[cut:]
}

// UpsertRecord adds or replaces a record in the history file by
// terminal ID.
func (store *TerminalHistoryStore) UpsertRecord(history TerminalHistoryFile, record TerminalHistoryRecord) (TerminalHistoryFile, error) {
	for i, existing := range history.Records {
		if existing.TerminalID == record.TerminalID {
			history.Records[i] = record
			return history, store.Save(history)
		}
	}
	history.Records = append(history.Records, record)
	return history, store.Save(history)
}

// DeleteRecord removes a record by terminal ID from the history file.
func (store *TerminalHistoryStore) DeleteRecord(history TerminalHistoryFile, terminalID string) (TerminalHistoryFile, error) {
	filtered := make([]TerminalHistoryRecord, 0, len(history.Records))
	for _, record := range history.Records {
		if record.TerminalID != terminalID {
			filtered = append(filtered, record)
		}
	}
	history.Records = filtered
	return history, store.Save(history)
}

// DeleteRecordsByProject removes all records for the given project ID.
func (store *TerminalHistoryStore) DeleteRecordsByProject(history TerminalHistoryFile, projectID string) (TerminalHistoryFile, error) {
	filtered := make([]TerminalHistoryRecord, 0, len(history.Records))
	for _, record := range history.Records {
		if record.ProjectID != projectID {
			filtered = append(filtered, record)
		}
	}
	history.Records = filtered
	return history, store.Save(history)
}

// DeleteRecordsByTodo removes all records for the given TODO ID.
func (store *TerminalHistoryStore) DeleteRecordsByTodo(history TerminalHistoryFile, todoID string) (TerminalHistoryFile, error) {
	filtered := make([]TerminalHistoryRecord, 0, len(history.Records))
	for _, record := range history.Records {
		if record.TodoID != todoID {
			filtered = append(filtered, record)
		}
	}
	history.Records = filtered
	return history, store.Save(history)
}

// DeleteRecordsByTodoProject removes all records for the given TODO
// project ID.
func (store *TerminalHistoryStore) DeleteRecordsByTodoProject(history TerminalHistoryFile, todoProjectID string) (TerminalHistoryFile, error) {
	filtered := make([]TerminalHistoryRecord, 0, len(history.Records))
	for _, record := range history.Records {
		if record.TodoProjectID != todoProjectID {
			filtered = append(filtered, record)
		}
	}
	history.Records = filtered
	return history, store.Save(history)
}
