package storage

import (
	"context"
	"database/sql"
	"fmt"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);`,
	`CREATE TABLE users(id TEXT PRIMARY KEY, name TEXT NOT NULL, role TEXT NOT NULL, password_hash TEXT NOT NULL, disabled INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL);
CREATE TABLE sessions(id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id), expires_at TEXT NOT NULL, revoked_at TEXT);
CREATE TABLE regions(id TEXT PRIMARY KEY, name TEXT NOT NULL, parent_id TEXT REFERENCES regions(id), flood_risk TEXT NOT NULL);
CREATE TABLE campaigns(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), name TEXT NOT NULL, status TEXT NOT NULL, start_at TEXT NOT NULL, end_at TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE plots(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), owner_id TEXT NOT NULL, acres REAL NOT NULL, crop TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE observations(id TEXT PRIMARY KEY, plot_id TEXT NOT NULL REFERENCES plots(id), observer_id TEXT NOT NULL REFERENCES users(id), kind TEXT NOT NULL, severity TEXT NOT NULL, notes TEXT NOT NULL, observed_at TEXT NOT NULL);
CREATE TABLE alerts(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), campaign_id TEXT REFERENCES campaigns(id), level TEXT NOT NULL, status TEXT NOT NULL, message TEXT NOT NULL, created_at TEXT NOT NULL, acknowledged_at TEXT);
CREATE TABLE field_tasks(id TEXT PRIMARY KEY, plot_id TEXT NOT NULL REFERENCES plots(id), assignee_id TEXT REFERENCES users(id), kind TEXT NOT NULL, status TEXT NOT NULL, lease_id TEXT, due_at TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE disaster_reports(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), reporter_id TEXT NOT NULL REFERENCES users(id), kind TEXT NOT NULL, status TEXT NOT NULL, description TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE recovery_cases(id TEXT PRIMARY KEY, report_id TEXT NOT NULL REFERENCES disaster_reports(id), region_id TEXT NOT NULL REFERENCES regions(id), status TEXT NOT NULL, owner_id TEXT REFERENCES users(id), estimated_loss REAL NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE drying_sites(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), name TEXT NOT NULL, capacity_tons REAL NOT NULL, active INTEGER NOT NULL DEFAULT 1, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE drying_reservations(id TEXT PRIMARY KEY, site_id TEXT NOT NULL REFERENCES drying_sites(id), plot_id TEXT NOT NULL REFERENCES plots(id), status TEXT NOT NULL, tons REAL NOT NULL, starts_at TEXT NOT NULL, ends_at TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE aid_cases(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), farmer_id TEXT NOT NULL, status TEXT NOT NULL, reason TEXT NOT NULL, amount_cents INTEGER NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE farm_projects(id TEXT PRIMARY KEY, region_id TEXT NOT NULL REFERENCES regions(id), name TEXT NOT NULL, status TEXT NOT NULL, budget_cents INTEGER NOT NULL, version INTEGER NOT NULL DEFAULT 1);
CREATE TABLE audit_events(id TEXT PRIMARY KEY, actor_id TEXT, action TEXT NOT NULL, object_type TEXT NOT NULL, object_id TEXT NOT NULL, result TEXT NOT NULL, request_id TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE outbox_events(id TEXT PRIMARY KEY, topic TEXT NOT NULL, aggregate_id TEXT NOT NULL, payload TEXT NOT NULL, status TEXT NOT NULL, attempts INTEGER NOT NULL DEFAULT 0, available_at TEXT NOT NULL);
CREATE TABLE idempotency_keys(key TEXT PRIMARY KEY, request_hash TEXT NOT NULL, response TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE INDEX idx_observations_plot_time ON observations(plot_id, observed_at);
CREATE INDEX idx_tasks_status_due ON field_tasks(status, due_at);
CREATE INDEX idx_reports_region_status ON disaster_reports(region_id, status);
CREATE INDEX idx_reservations_site_window ON drying_reservations(site_id, starts_at, ends_at);
CREATE INDEX idx_outbox_ready ON outbox_events(status, available_at);`,
}

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, migrations[0]); err != nil {
		return err
	}
	for version, script := range migrations[1:] {
		var applied int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version+1).Scan(&applied)
		if err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, script); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d: %w", version+1, err)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version, applied_at) VALUES (?, datetime('now'))", version+1); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
