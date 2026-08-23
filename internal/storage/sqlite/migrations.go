package sqlite

import (
	"context"
	"fmt"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);`,
	`CREATE TABLE IF NOT EXISTS regions(id TEXT PRIMARY KEY,name TEXT NOT NULL,parent_id TEXT,time_zone TEXT NOT NULL,coastal INTEGER NOT NULL,risk_level INTEGER NOT NULL,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(parent_id) REFERENCES regions(id)); CREATE INDEX IF NOT EXISTS idx_regions_parent ON regions(parent_id);`,
	`CREATE TABLE IF NOT EXISTS users(id TEXT PRIMARY KEY,username TEXT NOT NULL UNIQUE,password_hash TEXT NOT NULL,role TEXT NOT NULL,region_id TEXT,active INTEGER NOT NULL,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(region_id) REFERENCES regions(id)); CREATE TABLE IF NOT EXISTS sessions(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,token_hash TEXT NOT NULL UNIQUE,expires_at TEXT NOT NULL,revoked_at TEXT,created_at TEXT NOT NULL,FOREIGN KEY(user_id) REFERENCES users(id)); CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id,expires_at);`,
	`CREATE TABLE IF NOT EXISTS warnings(id TEXT PRIMARY KEY,typhoon_name TEXT NOT NULL,number TEXT NOT NULL,level TEXT NOT NULL,issued_at TEXT NOT NULL,effective_from TEXT NOT NULL,effective_until TEXT NOT NULL,rainfall_mm INTEGER NOT NULL,center_area TEXT NOT NULL,movement TEXT NOT NULL,status TEXT NOT NULL,version INTEGER NOT NULL); CREATE UNIQUE INDEX IF NOT EXISTS idx_warning_number_issued ON warnings(number,issued_at);`,
	`CREATE TABLE IF NOT EXISTS incidents(id TEXT PRIMARY KEY,warning_id TEXT NOT NULL,name TEXT NOT NULL,status TEXT NOT NULL,commander_id TEXT NOT NULL,activated_at TEXT NOT NULL,closed_at TEXT,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(warning_id) REFERENCES warnings(id),FOREIGN KEY(commander_id) REFERENCES users(id)); CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status,updated_at);`,
	`CREATE TABLE IF NOT EXISTS responses(id TEXT PRIMARY KEY,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,level TEXT NOT NULL,status TEXT NOT NULL,activated_by TEXT NOT NULL,activated_at TEXT NOT NULL,reason TEXT NOT NULL,version INTEGER NOT NULL,updated_at TEXT NOT NULL,UNIQUE(incident_id,region_id),FOREIGN KEY(incident_id) REFERENCES incidents(id),FOREIGN KEY(region_id) REFERENCES regions(id));`,
	`CREATE TABLE IF NOT EXISTS shelters(id TEXT PRIMARY KEY,region_id TEXT NOT NULL,name TEXT NOT NULL,capacity INTEGER NOT NULL,reserved INTEGER NOT NULL,occupied INTEGER NOT NULL,status TEXT NOT NULL,version INTEGER NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(region_id) REFERENCES regions(id)); CREATE TABLE IF NOT EXISTS reservations(id TEXT PRIMARY KEY,shelter_id TEXT NOT NULL,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,people INTEGER NOT NULL,status TEXT NOT NULL,idempotency_key TEXT NOT NULL,expires_at TEXT NOT NULL,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,UNIQUE(region_id,idempotency_key),FOREIGN KEY(shelter_id) REFERENCES shelters(id),FOREIGN KEY(incident_id) REFERENCES incidents(id));`,
	`CREATE TABLE IF NOT EXISTS teams(id TEXT PRIMARY KEY,region_id TEXT NOT NULL,name TEXT NOT NULL,specialty TEXT NOT NULL,status TEXT NOT NULL,current_incident_id TEXT,leader_id TEXT NOT NULL,version INTEGER NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(region_id) REFERENCES regions(id)); CREATE TABLE IF NOT EXISTS district_tasks(id TEXT PRIMARY KEY,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,kind TEXT NOT NULL,summary TEXT NOT NULL,status TEXT NOT NULL,priority INTEGER NOT NULL,assignee_id TEXT,created_by TEXT NOT NULL,deadline TEXT NOT NULL,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(incident_id) REFERENCES incidents(id),FOREIGN KEY(region_id) REFERENCES regions(id)); CREATE INDEX IF NOT EXISTS idx_tasks_filter ON district_tasks(incident_id,region_id,status,deadline);`,
	`CREATE TABLE IF NOT EXISTS dispatches(id TEXT PRIMARY KEY,team_id TEXT NOT NULL,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,task_id TEXT NOT NULL,status TEXT NOT NULL,requested_by TEXT NOT NULL,accepted_at TEXT,completed_at TEXT,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(team_id) REFERENCES teams(id),FOREIGN KEY(task_id) REFERENCES district_tasks(id)); CREATE INDEX IF NOT EXISTS idx_dispatch_team ON dispatches(team_id,status);`,
	`CREATE TABLE IF NOT EXISTS supply_lots(id TEXT PRIMARY KEY,region_id TEXT NOT NULL,kind TEXT NOT NULL,quantity INTEGER NOT NULL,reserved INTEGER NOT NULL,issued INTEGER NOT NULL,expires_at TEXT NOT NULL,status TEXT NOT NULL,version INTEGER NOT NULL,updated_at TEXT NOT NULL,FOREIGN KEY(region_id) REFERENCES regions(id)); CREATE TABLE IF NOT EXISTS allocations(id TEXT PRIMARY KEY,lot_id TEXT NOT NULL,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,task_id TEXT NOT NULL,quantity INTEGER NOT NULL,status TEXT NOT NULL,idempotency_key TEXT NOT NULL,version INTEGER NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL,UNIQUE(region_id,idempotency_key),FOREIGN KEY(lot_id) REFERENCES supply_lots(id),FOREIGN KEY(task_id) REFERENCES district_tasks(id));`,
	`CREATE TABLE IF NOT EXISTS receipts(id TEXT PRIMARY KEY,task_id TEXT NOT NULL,incident_id TEXT NOT NULL,region_id TEXT NOT NULL,reporter_id TEXT NOT NULL,summary TEXT NOT NULL,status TEXT NOT NULL,evidence_count INTEGER NOT NULL,submitted_at TEXT NOT NULL,reviewed_at TEXT,version INTEGER NOT NULL,FOREIGN KEY(task_id) REFERENCES district_tasks(id)); CREATE INDEX IF NOT EXISTS idx_receipts_incident ON receipts(incident_id,status);`,
	`CREATE TABLE IF NOT EXISTS audit_events(id TEXT PRIMARY KEY,actor_id TEXT NOT NULL,request_id TEXT NOT NULL,action TEXT NOT NULL,entity TEXT NOT NULL,entity_id TEXT NOT NULL,result TEXT NOT NULL,detail TEXT NOT NULL,created_at TEXT NOT NULL); CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_events(entity,entity_id,created_at);`,
	`CREATE TABLE IF NOT EXISTS outbox_events(id TEXT PRIMARY KEY,topic TEXT NOT NULL,aggregate_type TEXT NOT NULL,aggregate_id TEXT NOT NULL,payload BLOB NOT NULL,status TEXT NOT NULL,attempts INTEGER NOT NULL,next_attempt_at TEXT NOT NULL,last_error TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL); CREATE INDEX IF NOT EXISTS idx_outbox_due ON outbox_events(status,next_attempt_at);`,
	`CREATE TABLE IF NOT EXISTS jobs(id TEXT PRIMARY KEY,kind TEXT NOT NULL,aggregate_id TEXT NOT NULL,payload BLOB NOT NULL,status TEXT NOT NULL,attempts INTEGER NOT NULL,max_attempts INTEGER NOT NULL,next_run_at TEXT NOT NULL,last_error TEXT NOT NULL,locked_until TEXT,created_at TEXT NOT NULL,updated_at TEXT NOT NULL); CREATE INDEX IF NOT EXISTS idx_jobs_due ON jobs(status,next_run_at,locked_until); CREATE TABLE IF NOT EXISTS idempotency_records(scope TEXT NOT NULL,key TEXT NOT NULL,request_hash TEXT NOT NULL,status_code INTEGER NOT NULL,response BLOB NOT NULL,expires_at TEXT NOT NULL,created_at TEXT NOT NULL,PRIMARY KEY(scope,key));`,
}

func (s *Store) Migrate(ctx context.Context) error {
	for index, statement := range migrations {
		version := index + 1
		var exists int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&exists)
		if err != nil {
			return fmt.Errorf("inspect migration %d: %w", version, err)
		}
		if exists > 0 {
			var applied int
			if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version=?", version).Scan(&applied); err != nil {
				return fmt.Errorf("read migration %d: %w", version, err)
			}
			if applied > 0 {
				continue
			}
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version,applied_at) VALUES(?,datetime('now'))", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}
	return nil
}
