package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertRegion(ctx context.Context, v domain.Region) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO regions(id,name,parent_id,time_zone,coastal,risk_level,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.Name, nullText(v.ParentID), v.TimeZone, boolInt(v.Coastal), v.RiskLevel, v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "region", v.ID, err)
}
func (t *txStore) InsertShelter(ctx context.Context, v domain.Shelter) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO shelters(id,region_id,name,capacity,reserved,occupied,status,version,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.RegionID, v.Name, v.Capacity, v.Reserved, v.Occupied, v.Status, v.Version, stamp(v.UpdatedAt))
	return translate("insert", "shelter", v.ID, err)
}
func (t *txStore) InsertTeam(ctx context.Context, v domain.Team) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO teams(id,region_id,name,specialty,status,current_incident_id,leader_id,version,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.RegionID, v.Name, v.Specialty, v.Status, nullText(v.CurrentIncidentID), v.LeaderID, v.Version, stamp(v.UpdatedAt))
	return translate("insert", "team", v.ID, err)
}
func (t *txStore) InsertSupplyLot(ctx context.Context, v domain.SupplyLot) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO supply_lots(id,region_id,kind,quantity,reserved,issued,expires_at,status,version,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.ID, v.RegionID, v.Kind, v.Quantity, v.Reserved, v.Issued, stamp(v.ExpiresAt), v.Status, v.Version, stamp(v.UpdatedAt))
	return translate("insert", "supply_lot", v.ID, err)
}
