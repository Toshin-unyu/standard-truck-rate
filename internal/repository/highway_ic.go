package repository

import (
	"database/sql"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// HighwayICRepository 高速道路ICマスタのリポジトリ
type HighwayICRepository struct {
	db *sql.DB
}

// NewHighwayICRepository リポジトリを作成する
func NewHighwayICRepository(db *sql.DB) *HighwayICRepository {
	return &HighwayICRepository{db: db}
}

// Create ICマスタを1件作成する
func (r *HighwayICRepository) Create(ic *model.HighwayIC) error {
	_, err := r.db.Exec(`
		INSERT INTO highway_ic_master (code, name, yomi, type, road_no, road_name, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, ic.Code, ic.Name, ic.Yomi, ic.Type, ic.RoadNo, ic.RoadName, time.Now())
	return err
}

// BulkCreate ICマスタを一括作成する（トランザクション使用）
func (r *HighwayICRepository) BulkCreate(ics []*model.HighwayIC) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO highway_ic_master (code, name, yomi, type, road_no, road_name, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, ic := range ics {
		_, err := stmt.Exec(ic.Code, ic.Name, ic.Yomi, ic.Type, ic.RoadNo, ic.RoadName, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByCode コードでICを取得する
func (r *HighwayICRepository) GetByCode(code string) (*model.HighwayIC, error) {
	ic := &model.HighwayIC{}
	err := r.db.QueryRow(`
		SELECT code, name, yomi, type, road_no, road_name, updated_at
		FROM highway_ic_master WHERE code = ?
	`, code).Scan(&ic.Code, &ic.Name, &ic.Yomi, &ic.Type, &ic.RoadNo, &ic.RoadName, &ic.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ic, nil
}

// SearchByName 名前で部分一致検索する
func (r *HighwayICRepository) SearchByName(name string) ([]*model.HighwayIC, error) {
	rows, err := r.db.Query(`
		SELECT code, name, yomi, type, road_no, road_name, updated_at
		FROM highway_ic_master WHERE name LIKE ? ORDER BY name
	`, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighwayICs(rows)
}

// SearchByYomi 読みで前方一致検索する
func (r *HighwayICRepository) SearchByYomi(yomi string) ([]*model.HighwayIC, error) {
	rows, err := r.db.Query(`
		SELECT code, name, yomi, type, road_no, road_name, updated_at
		FROM highway_ic_master WHERE yomi LIKE ? ORDER BY yomi
	`, yomi+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighwayICs(rows)
}

// GetAll 全ICを取得する
func (r *HighwayICRepository) GetAll() ([]*model.HighwayIC, error) {
	rows, err := r.db.Query(`
		SELECT code, name, yomi, type, road_no, road_name, updated_at
		FROM highway_ic_master ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighwayICs(rows)
}

// Count IC件数を取得する
func (r *HighwayICRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM highway_ic_master`).Scan(&count)
	return count, err
}

// DeleteAll 全ICを削除する
func (r *HighwayICRepository) DeleteAll() error {
	_, err := r.db.Exec(`DELETE FROM highway_ic_master`)
	return err
}

// scanHighwayICs rowsからHighwayICスライスを作成する
func scanHighwayICs(rows *sql.Rows) ([]*model.HighwayIC, error) {
	var ics []*model.HighwayIC
	for rows.Next() {
		ic := &model.HighwayIC{}
		if err := rows.Scan(&ic.Code, &ic.Name, &ic.Yomi, &ic.Type, &ic.RoadNo, &ic.RoadName, &ic.UpdatedAt); err != nil {
			return nil, err
		}
		ics = append(ics, ic)
	}
	return ics, rows.Err()
}
