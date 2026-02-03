package repository

import (
	"database/sql"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// JtaTimeFareRepository JTA時間制運賃のリポジトリ
type JtaTimeFareRepository struct {
	db *sql.DB
}

// NewJtaTimeFareRepository リポジトリを作成する
func NewJtaTimeFareRepository(db *sql.DB) *JtaTimeFareRepository {
	return &JtaTimeFareRepository{db: db}
}

// === JtaTimeBaseFare (基礎額) ===

// CreateBaseFare 基礎額を作成する
func (r *JtaTimeFareRepository) CreateBaseFare(fare *model.JtaTimeBaseFare) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO jta_time_base_fares (region_code, vehicle_code, hours, base_km, fare_yen)
		VALUES (?, ?, ?, ?, ?)
	`, fare.RegionCode, fare.VehicleCode, fare.Hours, fare.BaseKm, fare.FareYen)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetBaseFareByID IDで基礎額を取得する
func (r *JtaTimeFareRepository) GetBaseFareByID(id int64) (*model.JtaTimeBaseFare, error) {
	fare := &model.JtaTimeBaseFare{}
	err := r.db.QueryRow(`
		SELECT id, region_code, vehicle_code, hours, base_km, fare_yen
		FROM jta_time_base_fares WHERE id = ?
	`, id).Scan(&fare.ID, &fare.RegionCode, &fare.VehicleCode, &fare.Hours, &fare.BaseKm, &fare.FareYen)
	if err != nil {
		return nil, err
	}
	return fare, nil
}

// GetAllBaseFares 全基礎額を取得する
func (r *JtaTimeFareRepository) GetAllBaseFares() ([]*model.JtaTimeBaseFare, error) {
	rows, err := r.db.Query(`
		SELECT id, region_code, vehicle_code, hours, base_km, fare_yen
		FROM jta_time_base_fares ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fares []*model.JtaTimeBaseFare
	for rows.Next() {
		fare := &model.JtaTimeBaseFare{}
		if err := rows.Scan(&fare.ID, &fare.RegionCode, &fare.VehicleCode, &fare.Hours, &fare.BaseKm, &fare.FareYen); err != nil {
			return nil, err
		}
		fares = append(fares, fare)
	}
	return fares, rows.Err()
}

// UpdateBaseFare 基礎額を更新する
func (r *JtaTimeFareRepository) UpdateBaseFare(fare *model.JtaTimeBaseFare) error {
	_, err := r.db.Exec(`
		UPDATE jta_time_base_fares
		SET region_code = ?, vehicle_code = ?, hours = ?, base_km = ?, fare_yen = ?
		WHERE id = ?
	`, fare.RegionCode, fare.VehicleCode, fare.Hours, fare.BaseKm, fare.FareYen, fare.ID)
	return err
}

// DeleteBaseFare 基礎額を削除する
func (r *JtaTimeFareRepository) DeleteBaseFare(id int64) error {
	_, err := r.db.Exec(`DELETE FROM jta_time_base_fares WHERE id = ?`, id)
	return err
}

// === JtaTimeSurcharge (加算額) ===

// CreateSurcharge 加算額を作成する
func (r *JtaTimeFareRepository) CreateSurcharge(surcharge *model.JtaTimeSurcharge) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO jta_time_surcharges (region_code, vehicle_code, surcharge_type, fare_yen)
		VALUES (?, ?, ?, ?)
	`, surcharge.RegionCode, surcharge.VehicleCode, surcharge.SurchargeType, surcharge.FareYen)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetSurchargeByID IDで加算額を取得する
func (r *JtaTimeFareRepository) GetSurchargeByID(id int64) (*model.JtaTimeSurcharge, error) {
	surcharge := &model.JtaTimeSurcharge{}
	err := r.db.QueryRow(`
		SELECT id, region_code, vehicle_code, surcharge_type, fare_yen
		FROM jta_time_surcharges WHERE id = ?
	`, id).Scan(&surcharge.ID, &surcharge.RegionCode, &surcharge.VehicleCode, &surcharge.SurchargeType, &surcharge.FareYen)
	if err != nil {
		return nil, err
	}
	return surcharge, nil
}

// GetAllSurcharges 全加算額を取得する
func (r *JtaTimeFareRepository) GetAllSurcharges() ([]*model.JtaTimeSurcharge, error) {
	rows, err := r.db.Query(`
		SELECT id, region_code, vehicle_code, surcharge_type, fare_yen
		FROM jta_time_surcharges ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var surcharges []*model.JtaTimeSurcharge
	for rows.Next() {
		s := &model.JtaTimeSurcharge{}
		if err := rows.Scan(&s.ID, &s.RegionCode, &s.VehicleCode, &s.SurchargeType, &s.FareYen); err != nil {
			return nil, err
		}
		surcharges = append(surcharges, s)
	}
	return surcharges, rows.Err()
}

// UpdateSurcharge 加算額を更新する
func (r *JtaTimeFareRepository) UpdateSurcharge(surcharge *model.JtaTimeSurcharge) error {
	_, err := r.db.Exec(`
		UPDATE jta_time_surcharges
		SET region_code = ?, vehicle_code = ?, surcharge_type = ?, fare_yen = ?
		WHERE id = ?
	`, surcharge.RegionCode, surcharge.VehicleCode, surcharge.SurchargeType, surcharge.FareYen, surcharge.ID)
	return err
}

// DeleteSurcharge 加算額を削除する
func (r *JtaTimeFareRepository) DeleteSurcharge(id int64) error {
	_, err := r.db.Exec(`DELETE FROM jta_time_surcharges WHERE id = ?`, id)
	return err
}

// === TimeFareGetter インターフェース実装 ===

// GetBaseFare 運輸局・車格・時間制で基礎額を取得（TimeFareGetterインターフェース実装）
func (r *JtaTimeFareRepository) GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error) {
	fare := &model.JtaTimeBaseFare{}
	err := r.db.QueryRow(`
		SELECT id, region_code, vehicle_code, hours, base_km, fare_yen
		FROM jta_time_base_fares
		WHERE region_code = ? AND vehicle_code = ? AND hours = ?
	`, regionCode, vehicleCode, hours).Scan(&fare.ID, &fare.RegionCode, &fare.VehicleCode, &fare.Hours, &fare.BaseKm, &fare.FareYen)
	if err != nil {
		return nil, err
	}
	return fare, nil
}

// GetSurcharge 運輸局・車格・種別で加算額を取得（TimeFareGetterインターフェース実装）
func (r *JtaTimeFareRepository) GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error) {
	surcharge := &model.JtaTimeSurcharge{}
	err := r.db.QueryRow(`
		SELECT id, region_code, vehicle_code, surcharge_type, fare_yen
		FROM jta_time_surcharges
		WHERE region_code = ? AND vehicle_code = ? AND surcharge_type = ?
	`, regionCode, vehicleCode, surchargeType).Scan(&surcharge.ID, &surcharge.RegionCode, &surcharge.VehicleCode, &surcharge.SurchargeType, &surcharge.FareYen)
	if err != nil {
		return nil, err
	}
	return surcharge, nil
}
