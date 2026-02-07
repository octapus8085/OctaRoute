package routing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type StateStore struct {
	db *sql.DB
}

func OpenStateStore(path string) (*StateStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open routing state db: %w", err)
	}
	store := &StateStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *StateStore) Close() error {
	return s.db.Close()
}

func (s *StateStore) migrate(ctx context.Context) error {
	schema := `
    CREATE TABLE IF NOT EXISTS routing_state (
        id INTEGER PRIMARY KEY CHECK (id = 1),
        payload TEXT NOT NULL,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    `
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply routing_state schema: %w", err)
	}
	return nil
}

func (s *StateStore) Save(ctx context.Context, state RoutingState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal routing state: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
        INSERT INTO routing_state (id, payload, updated_at)
        VALUES (1, ?, ?)
        ON CONFLICT(id) DO UPDATE SET payload = excluded.payload, updated_at = excluded.updated_at
    `, string(payload), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save routing state: %w", err)
	}
	return nil
}

func (s *StateStore) Load(ctx context.Context) (RoutingState, bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT payload FROM routing_state WHERE id = 1`)
	var payload string
	if err := row.Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return RoutingState{}, false, nil
		}
		return RoutingState{}, false, fmt.Errorf("load routing state: %w", err)
	}
	var state RoutingState
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return RoutingState{}, false, fmt.Errorf("unmarshal routing state: %w", err)
	}
	return state, true, nil
}
