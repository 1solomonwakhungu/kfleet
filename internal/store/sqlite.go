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
	agent_version TEXT NOT NULL DEFAULT '',
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

const createSnapshotNodesTable = `
CREATE TABLE IF NOT EXISTS snapshot_nodes (
	cluster_id TEXT NOT NULL,
	name TEXT NOT NULL,
	status TEXT NOT NULL,
	roles TEXT NOT NULL,
	version TEXT NOT NULL,
	cpu_capacity TEXT NOT NULL,
	memory_capacity TEXT NOT NULL,
	ready INTEGER NOT NULL CHECK (ready IN (0, 1)),
	PRIMARY KEY (cluster_id, name),
	FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
)`

const createSnapshotPodsTable = `
CREATE TABLE IF NOT EXISTS snapshot_pods (
	cluster_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	name TEXT NOT NULL,
	phase TEXT NOT NULL,
	node_name TEXT NOT NULL,
	restart_count INTEGER NOT NULL,
	ready INTEGER NOT NULL CHECK (ready IN (0, 1)),
	start_time TIMESTAMP NOT NULL,
	PRIMARY KEY (cluster_id, namespace, name),
	FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
)`

const createSnapshotServicesTable = `
CREATE TABLE IF NOT EXISTS snapshot_services (
	cluster_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	cluster_ip TEXT NOT NULL,
	external_ips TEXT NOT NULL,
	ports TEXT NOT NULL,
	age TEXT NOT NULL,
	PRIMARY KEY (cluster_id, namespace, name),
	FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
)`

const createSnapshotDeploymentsTable = `
CREATE TABLE IF NOT EXISTS snapshot_deployments (
	cluster_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	name TEXT NOT NULL,
	ready_replicas INTEGER NOT NULL,
	desired_replicas INTEGER NOT NULL,
	updated_replicas INTEGER NOT NULL,
	available_replicas INTEGER NOT NULL,
	age TEXT NOT NULL,
	PRIMARY KEY (cluster_id, namespace, name),
	FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
)`

const createSnapshotEventsTable = `
CREATE TABLE IF NOT EXISTS snapshot_events (
	cluster_id TEXT NOT NULL,
	sequence INTEGER NOT NULL,
	namespace TEXT NOT NULL,
	reason TEXT NOT NULL,
	message TEXT NOT NULL,
	type TEXT NOT NULL,
	count INTEGER NOT NULL,
	last_timestamp TIMESTAMP NOT NULL,
	PRIMARY KEY (cluster_id, sequence),
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

	// Keep PRAGMA foreign_keys effective for every operation by using one
	// connection. SQLite serializes writers regardless, so this also avoids
	// surprising per-connection pragma behavior.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to sqlite database: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	for _, migration := range []struct {
		name string
		sql  string
	}{
		{name: "clusters", sql: createClustersTable},
		{name: "agents", sql: createAgentsTable},
		{name: "snapshot nodes", sql: createSnapshotNodesTable},
		{name: "snapshot pods", sql: createSnapshotPodsTable},
		{name: "snapshot services", sql: createSnapshotServicesTable},
		{name: "snapshot deployments", sql: createSnapshotDeploymentsTable},
		{name: "snapshot events", sql: createSnapshotEventsTable},
		{name: "alert rules", sql: createAlertRulesTable},
		{name: "alerts", sql: createAlertsTable},
		{name: "alert history index", sql: createAlertHistoryIndex},
		{name: "alert delivery index", sql: createAlertDeliveryIndex},
	} {
		if _, err := db.Exec(migration.sql); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("migrate %s table: %w", migration.name, err)
		}
	}
	if err := ensureColumn(db, "clusters", "agent_version", "TEXT NOT NULL DEFAULT ''"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate clusters agent_version column: %w", err)
	}
	if err := seedDefaultAlertRules(db, time.Now().UTC()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seed default alert rules: %w", err)
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
			id, name, health, version, agent_version, node_count, pod_count,
			registered_at, last_heartbeat, labels
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cluster.ID,
		cluster.Name,
		cluster.Health,
		cluster.Version,
		cluster.AgentVersion,
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
		SELECT id, name, health, version, agent_version, node_count, pod_count,
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
		SELECT id, name, health, version, agent_version, node_count, pod_count,
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
		SELECT c.id, c.name, c.health, c.version, c.agent_version, c.node_count, c.pod_count,
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

func (s *sqliteStore) ReplaceSnapshot(
	ctx context.Context,
	id string,
	snapshot types.ClusterSnapshot,
	version, agentVersion string,
	health types.ClusterHealth,
	lastHeartbeat time.Time,
) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin snapshot transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		UPDATE clusters
		SET node_count = ?, pod_count = ?, version = ?, agent_version = ?, health = ?, last_heartbeat = ?
		WHERE id = ?`, len(snapshot.Nodes), len(snapshot.Pods), version, agentVersion, health, lastHeartbeat, id)
	if err != nil {
		return fmt.Errorf("update snapshot metadata: %w", err)
	}
	if err := requireAffectedRow(result); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agents SET last_heartbeat = ? WHERE cluster_id = ?`, lastHeartbeat, id); err != nil {
		return fmt.Errorf("update snapshot heartbeat: %w", err)
	}

	for _, table := range []string{
		"snapshot_nodes",
		"snapshot_pods",
		"snapshot_services",
		"snapshot_deployments",
		"snapshot_events",
	} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE cluster_id = ?", id); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	for _, node := range snapshot.Nodes {
		roles, marshalErr := json.Marshal(nonNil(node.Roles))
		if marshalErr != nil {
			return fmt.Errorf("encode node roles: %w", marshalErr)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO snapshot_nodes (
				cluster_id, name, status, roles, version, cpu_capacity, memory_capacity, ready
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, node.Name, node.Status, string(roles), node.Version,
			node.CPUCapacity, node.MemoryCapacity, boolInt(node.Ready)); err != nil {
			return fmt.Errorf("insert snapshot node: %w", err)
		}
	}
	for _, pod := range snapshot.Pods {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO snapshot_pods (
				cluster_id, namespace, name, phase, node_name, restart_count, ready, start_time
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, pod.Namespace, pod.Name, pod.Phase, pod.NodeName,
			pod.RestartCount, boolInt(pod.Ready), pod.StartTime); err != nil {
			return fmt.Errorf("insert snapshot pod: %w", err)
		}
	}
	for _, service := range snapshot.Services {
		externalIPs, marshalErr := json.Marshal(nonNil(service.ExternalIPs))
		if marshalErr != nil {
			return fmt.Errorf("encode service external IPs: %w", marshalErr)
		}
		ports, marshalErr := json.Marshal(nonNil(service.Ports))
		if marshalErr != nil {
			return fmt.Errorf("encode service ports: %w", marshalErr)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO snapshot_services (
				cluster_id, namespace, name, type, cluster_ip, external_ips, ports, age
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, service.Namespace, service.Name, service.Type, service.ClusterIP,
			string(externalIPs), string(ports), service.Age); err != nil {
			return fmt.Errorf("insert snapshot service: %w", err)
		}
	}
	for _, deployment := range snapshot.Deployments {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO snapshot_deployments (
				cluster_id, namespace, name, ready_replicas, desired_replicas,
				updated_replicas, available_replicas, age
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, deployment.Namespace, deployment.Name, deployment.ReadyReplicas,
			deployment.DesiredReplicas, deployment.UpdatedReplicas,
			deployment.AvailableReplicas, deployment.Age); err != nil {
			return fmt.Errorf("insert snapshot deployment: %w", err)
		}
	}
	for index, event := range snapshot.Events {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO snapshot_events (
				cluster_id, sequence, namespace, reason, message, type, count, last_timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, index, event.Namespace, event.Reason, event.Message,
			event.Type, event.Count, event.LastTimestamp); err != nil {
			return fmt.Errorf("insert snapshot event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit snapshot transaction: %w", err)
	}
	return nil
}

func (s *sqliteStore) ListNodes(ctx context.Context, clusterID string) ([]types.Node, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, status, roles, version, cpu_capacity, memory_capacity, ready
		FROM snapshot_nodes WHERE cluster_id = ? ORDER BY name`, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list snapshot nodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]types.Node, 0)
	for rows.Next() {
		var node types.Node
		var roles string
		var ready int
		if err := rows.Scan(&node.Name, &node.Status, &roles, &node.Version, &node.CPUCapacity, &node.MemoryCapacity, &ready); err != nil {
			return nil, fmt.Errorf("scan snapshot node: %w", err)
		}
		if err := json.Unmarshal([]byte(roles), &node.Roles); err != nil {
			return nil, fmt.Errorf("decode snapshot node roles: %w", err)
		}
		node.Ready = ready == 1
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot nodes: %w", err)
	}
	return nodes, nil
}

func (s *sqliteStore) ListPods(ctx context.Context, clusterID, namespace string) ([]types.Pod, error) {
	query := `
		SELECT name, namespace, phase, node_name, restart_count, ready, start_time
		FROM snapshot_pods WHERE cluster_id = ?`
	args := []any{clusterID}
	if namespace != "" {
		query += " AND namespace = ?"
		args = append(args, namespace)
	}
	query += " ORDER BY namespace, name"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snapshot pods: %w", err)
	}
	defer rows.Close()

	pods := make([]types.Pod, 0)
	for rows.Next() {
		var pod types.Pod
		var ready int
		if err := rows.Scan(&pod.Name, &pod.Namespace, &pod.Phase, &pod.NodeName, &pod.RestartCount, &ready, &pod.StartTime); err != nil {
			return nil, fmt.Errorf("scan snapshot pod: %w", err)
		}
		pod.Ready = ready == 1
		pods = append(pods, pod)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot pods: %w", err)
	}
	return pods, nil
}

func (s *sqliteStore) ListServices(ctx context.Context, clusterID, namespace string) ([]types.Service, error) {
	query := `
		SELECT name, namespace, type, cluster_ip, external_ips, ports, age
		FROM snapshot_services WHERE cluster_id = ?`
	args := []any{clusterID}
	if namespace != "" {
		query += " AND namespace = ?"
		args = append(args, namespace)
	}
	query += " ORDER BY namespace, name"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snapshot services: %w", err)
	}
	defer rows.Close()

	services := make([]types.Service, 0)
	for rows.Next() {
		var service types.Service
		var externalIPs, ports string
		if err := rows.Scan(&service.Name, &service.Namespace, &service.Type, &service.ClusterIP, &externalIPs, &ports, &service.Age); err != nil {
			return nil, fmt.Errorf("scan snapshot service: %w", err)
		}
		if err := json.Unmarshal([]byte(externalIPs), &service.ExternalIPs); err != nil {
			return nil, fmt.Errorf("decode service external IPs: %w", err)
		}
		if err := json.Unmarshal([]byte(ports), &service.Ports); err != nil {
			return nil, fmt.Errorf("decode service ports: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot services: %w", err)
	}
	return services, nil
}

func (s *sqliteStore) ListDeployments(ctx context.Context, clusterID, namespace string) ([]types.Deployment, error) {
	query := `
		SELECT name, namespace, ready_replicas, desired_replicas,
		       updated_replicas, available_replicas, age
		FROM snapshot_deployments WHERE cluster_id = ?`
	args := []any{clusterID}
	if namespace != "" {
		query += " AND namespace = ?"
		args = append(args, namespace)
	}
	query += " ORDER BY namespace, name"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snapshot deployments: %w", err)
	}
	defer rows.Close()

	deployments := make([]types.Deployment, 0)
	for rows.Next() {
		var deployment types.Deployment
		if err := rows.Scan(
			&deployment.Name,
			&deployment.Namespace,
			&deployment.ReadyReplicas,
			&deployment.DesiredReplicas,
			&deployment.UpdatedReplicas,
			&deployment.AvailableReplicas,
			&deployment.Age,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot deployment: %w", err)
		}
		deployments = append(deployments, deployment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot deployments: %w", err)
	}
	return deployments, nil
}

func (s *sqliteStore) ListEvents(ctx context.Context, clusterID, namespace string) ([]types.Event, error) {
	query := `
		SELECT namespace, reason, message, type, count, last_timestamp
		FROM snapshot_events WHERE cluster_id = ?`
	args := []any{clusterID}
	if namespace != "" {
		query += " AND namespace = ?"
		args = append(args, namespace)
	}
	query += " ORDER BY sequence"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snapshot events: %w", err)
	}
	defer rows.Close()

	events := make([]types.Event, 0)
	for rows.Next() {
		event := types.Event{ClusterID: clusterID}
		if err := rows.Scan(&event.Namespace, &event.Reason, &event.Message, &event.Type, &event.Count, &event.LastTimestamp); err != nil {
			return nil, fmt.Errorf("scan snapshot event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot events: %w", err)
	}
	return events, nil
}

func (s *sqliteStore) ListNamespaces(ctx context.Context, clusterID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT namespace FROM (
			SELECT namespace FROM snapshot_pods WHERE cluster_id = ?
			UNION SELECT namespace FROM snapshot_services WHERE cluster_id = ?
			UNION SELECT namespace FROM snapshot_deployments WHERE cluster_id = ?
			UNION SELECT namespace FROM snapshot_events WHERE cluster_id = ?
		) WHERE namespace != '' ORDER BY namespace`, clusterID, clusterID, clusterID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list snapshot namespaces: %w", err)
	}
	defer rows.Close()

	namespaces := make([]string, 0)
	for rows.Next() {
		var namespace string
		if err := rows.Scan(&namespace); err != nil {
			return nil, fmt.Errorf("scan snapshot namespace: %w", err)
		}
		namespaces = append(namespaces, namespace)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list snapshot namespaces: %w", err)
	}
	return namespaces, nil
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
		&cluster.AgentVersion,
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

func ensureColumn(db *sql.DB, table, column, definition string) error {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			_ = rows.Close()
			return err
		}
		if name == column {
			found = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if found {
		return nil
	}
	_, err = db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + definition)
	return err
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

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nonNil[T any](values []T) []T {
	if values == nil {
		return make([]T, 0)
	}
	return values
}
