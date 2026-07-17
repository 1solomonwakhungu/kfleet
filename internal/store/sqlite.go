package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	_ "modernc.org/sqlite"
)

const createClustersTable = `
CREATE TABLE IF NOT EXISTS clusters (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	health TEXT NOT NULL,
	version TEXT,
	node_count INTEGER,
	pod_count INTEGER,
	registered_at TIMESTAMP,
	last_heartbeat TIMESTAMP,
	labels TEXT
)`

type sqliteStore struct {
	db *sql.DB
}

// Open opens a SQLite database at dbPath and applies its schema migrations.
func Open(dbPath string) (Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to sqlite database: %w", err)
	}
	if _, err := db.Exec(createClustersTable); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite database: %w", err)
	}

	return &sqliteStore{db: db}, nil
}

func (s *sqliteStore) CreateCluster(ctx context.Context, cluster types.Cluster) error {
	labels, err := json.Marshal(cluster.Labels)
	if err != nil {
		return fmt.Errorf("encode cluster labels: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO clusters (
			id, name, health, version, node_count, pod_count,
			registered_at, last_heartbeat, labels
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cluster.ID,
		cluster.Name,
		cluster.Health,
		cluster.Version,
		cluster.NodeCount,
		cluster.PodCount,
		cluster.RegisteredAt,
		cluster.LastHeartbeat,
		string(labels),
	)
	if err != nil {
		return fmt.Errorf("create cluster: %w", err)
	}
	return nil
}

func (s *sqliteStore) GetCluster(ctx context.Context, id string) (types.Cluster, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, health, version, node_count, pod_count,
		       registered_at, last_heartbeat, labels
		FROM clusters
		WHERE id = ?`, id)

	cluster, err := scanCluster(row)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Cluster{}, ErrNotFound
	}
	if err != nil {
		return types.Cluster{}, fmt.Errorf("get cluster: %w", err)
	}
	return cluster, nil
}

func (s *sqliteStore) ListClusters(ctx context.Context) ([]types.Cluster, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, health, version, node_count, pod_count,
		       registered_at, last_heartbeat, labels
		FROM clusters
		ORDER BY registered_at, name`)
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	defer rows.Close()

	clusters := make([]types.Cluster, 0)
	for rows.Next() {
		cluster, err := scanCluster(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cluster: %w", err)
		}
		clusters = append(clusters, cluster)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	return clusters, nil
}

func (s *sqliteStore) DeleteCluster(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM clusters WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) UpdateHealth(
	ctx context.Context,
	id string,
	health types.ClusterHealth,
	lastHeartbeat time.Time,
) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE clusters
		SET health = ?, last_heartbeat = ?
		WHERE id = ?`, health, lastHeartbeat, id)
	if err != nil {
		return fmt.Errorf("update cluster health: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) UpdateSnapshot(
	ctx context.Context,
	id string,
	nodeCount, podCount int,
	version string,
) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE clusters
		SET node_count = ?, pod_count = ?, version = ?
		WHERE id = ?`, nodeCount, podCount, version, id)
	if err != nil {
		return fmt.Errorf("update cluster snapshot: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

type clusterScanner interface {
	Scan(dest ...any) error
}

func scanCluster(scanner clusterScanner) (types.Cluster, error) {
	var cluster types.Cluster
	var labels string
	if err := scanner.Scan(
		&cluster.ID,
		&cluster.Name,
		&cluster.Health,
		&cluster.Version,
		&cluster.NodeCount,
		&cluster.PodCount,
		&cluster.RegisteredAt,
		&cluster.LastHeartbeat,
		&labels,
	); err != nil {
		return types.Cluster{}, err
	}
	if err := json.Unmarshal([]byte(labels), &cluster.Labels); err != nil {
		return types.Cluster{}, fmt.Errorf("decode cluster labels: %w", err)
	}
	return cluster, nil
}

func requireAffectedRow(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
