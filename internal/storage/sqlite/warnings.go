package sqlite

import (
	"context"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertWarning(ctx context.Context, v domain.Warning) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO warnings(id,typhoon_name,number,level,issued_at,effective_from,effective_until,rainfall_mm,center_area,movement,status,version) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.TyphoonName, v.Number, v.Level, stamp(v.IssuedAt), stamp(v.EffectiveFrom), stamp(v.EffectiveUntil), v.RainfallMM, v.CenterArea, v.Movement, v.Status, v.Version)
	return translate("insert", "warning", v.ID, err)
}
func (t *txStore) GetWarning(ctx context.Context, id string) (domain.Warning, error) {
	var v domain.Warning
	var issued, from, until string
	err := t.tx.QueryRowContext(ctx, `SELECT id,typhoon_name,number,level,issued_at,effective_from,effective_until,rainfall_mm,center_area,movement,status,version FROM warnings WHERE id=?`, id).Scan(&v.ID, &v.TyphoonName, &v.Number, &v.Level, &issued, &from, &until, &v.RainfallMM, &v.CenterArea, &v.Movement, &v.Status, &v.Version)
	if err != nil {
		return domain.Warning{}, translate("get", "warning", id, err)
	}
	v.IssuedAt = parseStamp(issued)
	v.EffectiveFrom = parseStamp(from)
	v.EffectiveUntil = parseStamp(until)
	return v, nil
}
func (t *txStore) UpdateWarning(ctx context.Context, v domain.Warning, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE warnings SET level=?,effective_from=?,effective_until=?,rainfall_mm=?,center_area=?,movement=?,status=?,version=version+1 WHERE id=? AND version=?`, v.Level, stamp(v.EffectiveFrom), stamp(v.EffectiveUntil), v.RainfallMM, v.CenterArea, v.Movement, v.Status, v.ID, expected)
	if err != nil {
		return translate("update", "warning", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("warning", v.ID, expected, v.Version)
	}
	return nil
}
