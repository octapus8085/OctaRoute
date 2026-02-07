package controllerdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

type Node struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Zone      string    `json:"zone"`
	CreatedAt time.Time `json:"createdAt"`
}

type Policy struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Route struct {
	ID        int64     `json:"id"`
	CIDR      string    `json:"cidr"`
	NextHop   string    `json:"nextHop"`
	NodeID    int64     `json:"nodeId"`
	CreatedAt time.Time `json:"createdAt"`
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	schema := `
    PRAGMA foreign_keys = ON;
    CREATE TABLE IF NOT EXISTS nodes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        address TEXT NOT NULL,
        zone TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS policies (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        source TEXT NOT NULL,
        destination TEXT NOT NULL,
        action TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS routes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cidr TEXT NOT NULL,
        next_hop TEXT NOT NULL,
        node_id INTEGER NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(node_id) REFERENCES nodes(id) ON DELETE CASCADE
    );
    `
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

func (s *Store) ListNodes(ctx context.Context) ([]Node, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, address, zone, created_at FROM nodes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Address, &n.Zone, &n.CreatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (s *Store) CreateNode(ctx context.Context, n Node) (Node, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO nodes (name, address, zone) VALUES (?, ?, ?)`, n.Name, n.Address, n.Zone)
	if err != nil {
		return Node{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Node{}, err
	}
	n.ID = id
	return n, nil
}

func (s *Store) ListPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, source, destination, action, created_at FROM policies ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Name, &p.Source, &p.Destination, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (s *Store) CreatePolicy(ctx context.Context, p Policy) (Policy, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO policies (name, source, destination, action) VALUES (?, ?, ?, ?)`, p.Name, p.Source, p.Destination, p.Action)
	if err != nil {
		return Policy{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Policy{}, err
	}
	p.ID = id
	return p, nil
}

func (s *Store) ListRoutes(ctx context.Context) ([]Route, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, cidr, next_hop, node_id, created_at FROM routes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var r Route
		if err := rows.Scan(&r.ID, &r.CIDR, &r.NextHop, &r.NodeID, &r.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}
	return routes, rows.Err()
}

func (s *Store) CreateRoute(ctx context.Context, r Route) (Route, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO routes (cidr, next_hop, node_id) VALUES (?, ?, ?)`, r.CIDR, r.NextHop, r.NodeID)
	if err != nil {
		return Route{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Route{}, err
	}
	r.ID = id
	return r, nil
}
