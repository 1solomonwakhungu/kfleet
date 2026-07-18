package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	_ "modernc.org/sqlite" // pure-Go SQLite driver; registers with database/sql for CGO-free builds
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

const createAgentsTable = `
CREATE TABLE IF NOT EXISTS agents (
	cluster_id TEXT PRIMARY KEY,
	token_hash TEXT NOT NULL,
	approved INTEGER NOT NULL DEFAULT 0 CHECK (approved IN (0, 1)),
	issued_at TIMESTAMP NOT NULL,
	last_heartbeat TIMESTAMP,
	FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
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
	if _, err := db.Exec(createAgentsTable); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate agents table: %w", err)
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
	if err := requireAffectedRow(result); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE agents SET last_heartbeat = ? WHERE cluster_id = ?`, lastHeartbeat, id); err != nil {
		return fmt.Errorf("update agent heartbeat: %w", err)
	}
	return nil
}

func (s *sqliteStore) IssueAgentToken(ctx context.Context, clusterID, tokenHash string) error {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (cluster_id, token_hash, approved, issued_at, last_heartbeat)
		SELECT id, ?, 0, ?, NULL FROM clusters WHERE id = ?
		ON CONFLICT(cluster_id) DO UPDATE SET
			token_hash = excluded.token_hash,
			approved = 0,
			issued_at = excluded.issued_at,
			last_heartbeat = NULL`, tokenHash, time.Now().UTC(), clusterID)
	if err != nil {
		return fmt.Errorf("issue agent token: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) ValidateAgentToken(ctx context.Context, clusterID, tokenHash string) (bool, error) {
	var approved int
	err := s.db.QueryRowContext(ctx, `
		SELECT approved FROM agents WHERE cluster_id = ? AND token_hash = ?`, clusterID, tokenHash).Scan(&approved)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf("validate agent token: %w", err)
	}
	return approved == 1, nil
}

func (s *sqliteStore) ApproveAgent(ctx context.Context, clusterID string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE agents SET approved = 1 WHERE cluster_id = ?`, clusterID)
	if err != nil {
		return fmt.Errorf("approve agent: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) ListPendingAgents(ctx context.Context) ([]types.Cluster, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.name, c.health, c.version, c.node_count, c.pod_count,
		       c.registered_at, c.last_heartbeat, c.labels
		FROM clusters c
		JOIN agents a ON a.cluster_id = c.id
		WHERE a.approved = 0
		ORDER BY a.issued_at, c.name`)
	if err != nil {
		return nil, fmt.Errorf("list pending agents: %w", err)
	}
	defer rows.Close()

	clusters := make([]types.Cluster, 0)
	for rows.Next() {
		cluster, err := scanCluster(rows)
		if err != nil {
			return nil, fmt.Errorf("scan pending agent: %w", err)
		}
		clusters = append(clusters, cluster)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list pending agents: %w", err)
	}
	return clusters, nil
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
