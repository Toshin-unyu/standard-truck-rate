package repository

import (
	"database/sql"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// AkabouFareRepository 赤帽運賃のリポジトリ
type AkabouFareRepository struct {
	db *sql.DB
}

// NewAkabouFareRepository リポジトリを作成する
func NewAkabouFareRepository(db *sql.DB) *AkabouFareRepository {
	return &AkabouFareRepository{db: db}
}

// === AkabouDistanceFare (距離制運賃) ===

// CreateDistanceFare 距離制運賃を作成する
func (r *AkabouFareRepository) CreateDistanceFare(fare *model.AkabouDistanceFare) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO akabou_distance_fares (min_km, max_km, base_fare, per_km_rate)
		VALUES (?, ?, ?, ?)
	`, fare.MinKm, fare.MaxKm, fare.BaseFare, fare.PerKmRate)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetDistanceFareByID IDで距離制運賃を取得する
func (r *AkabouFareRepository) GetDistanceFareByID(id int64) (*model.AkabouDistanceFare, error) {
	fare := &model.AkabouDistanceFare{}
	err := r.db.QueryRow(`
		SELECT id, min_km, max_km, base_fare, per_km_rate
		FROM akabou_distance_fares WHERE id = ?
	`, id).Scan(&fare.ID, &fare.MinKm, &fare.MaxKm, &fare.BaseFare, &fare.PerKmRate)
	if err != nil {
		return nil, err
	}
	return fare, nil
}

// GetAllDistanceFares 全距離制運賃を取得する
func (r *AkabouFareRepository) GetAllDistanceFares() ([]*model.AkabouDistanceFare, error) {
	rows, err := r.db.Query(`
		SELECT id, min_km, max_km, base_fare, per_km_rate
		FROM akabou_distance_fares ORDER BY min_km
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fares []*model.AkabouDistanceFare
	for rows.Next() {
		fare := &model.AkabouDistanceFare{}
		if err := rows.Scan(&fare.ID, &fare.MinKm, &fare.MaxKm, &fare.BaseFare, &fare.PerKmRate); err != nil {
			return nil, err
		}
		fares = append(fares, fare)
	}
	return fares, rows.Err()
}

// UpdateDistanceFare 距離制運賃を更新する
func (r *AkabouFareRepository) UpdateDistanceFare(fare *model.AkabouDistanceFare) error {
	_, err := r.db.Exec(`
		UPDATE akabou_distance_fares
		SET min_km = ?, max_km = ?, base_fare = ?, per_km_rate = ?
		WHERE id = ?
	`, fare.MinKm, fare.MaxKm, fare.BaseFare, fare.PerKmRate, fare.ID)
	return err
}

// DeleteDistanceFare 距離制運賃を削除する
func (r *AkabouFareRepository) DeleteDistanceFare(id int64) error {
	_, err := r.db.Exec(`DELETE FROM akabou_distance_fares WHERE id = ?`, id)
	return err
}

// === AkabouTimeFare (時間制運賃) ===

// CreateTimeFare 時間制運賃を作成する
func (r *AkabouFareRepository) CreateTimeFare(fare *model.AkabouTimeFare) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO akabou_time_fares (base_hours, base_km, base_fare, overtime_rate)
		VALUES (?, ?, ?, ?)
	`, fare.BaseHours, fare.BaseKm, fare.BaseFare, fare.OvertimeRate)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetTimeFareByID IDで時間制運賃を取得する
func (r *AkabouFareRepository) GetTimeFareByID(id int64) (*model.AkabouTimeFare, error) {
	fare := &model.AkabouTimeFare{}
	err := r.db.QueryRow(`
		SELECT id, base_hours, base_km, base_fare, overtime_rate
		FROM akabou_time_fares WHERE id = ?
	`, id).Scan(&fare.ID, &fare.BaseHours, &fare.BaseKm, &fare.BaseFare, &fare.OvertimeRate)
	if err != nil {
		return nil, err
	}
	return fare, nil
}

// GetAllTimeFares 全時間制運賃を取得する
func (r *AkabouFareRepository) GetAllTimeFares() ([]*model.AkabouTimeFare, error) {
	rows, err := r.db.Query(`
		SELECT id, base_hours, base_km, base_fare, overtime_rate
		FROM akabou_time_fares ORDER BY base_hours
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fares []*model.AkabouTimeFare
	for rows.Next() {
		fare := &model.AkabouTimeFare{}
		if err := rows.Scan(&fare.ID, &fare.BaseHours, &fare.BaseKm, &fare.BaseFare, &fare.OvertimeRate); err != nil {
			return nil, err
		}
		fares = append(fares, fare)
	}
	return fares, rows.Err()
}

// UpdateTimeFare 時間制運賃を更新する
func (r *AkabouFareRepository) UpdateTimeFare(fare *model.AkabouTimeFare) error {
	_, err := r.db.Exec(`
		UPDATE akabou_time_fares
		SET base_hours = ?, base_km = ?, base_fare = ?, overtime_rate = ?
		WHERE id = ?
	`, fare.BaseHours, fare.BaseKm, fare.BaseFare, fare.OvertimeRate, fare.ID)
	return err
}

// DeleteTimeFare 時間制運賃を削除する
func (r *AkabouFareRepository) DeleteTimeFare(id int64) error {
	_, err := r.db.Exec(`DELETE FROM akabou_time_fares WHERE id = ?`, id)
	return err
}

// === AkabouSurcharge (割増料金) ===

// CreateSurcharge 割増料金を作成する
func (r *AkabouFareRepository) CreateSurcharge(surcharge *model.AkabouSurcharge) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO akabou_surcharges (surcharge_type, rate_percent, description)
		VALUES (?, ?, ?)
	`, surcharge.SurchargeType, surcharge.RatePercent, surcharge.Description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetSurchargeByID IDで割増料金を取得する
func (r *AkabouFareRepository) GetSurchargeByID(id int64) (*model.AkabouSurcharge, error) {
	surcharge := &model.AkabouSurcharge{}
	err := r.db.QueryRow(`
		SELECT id, surcharge_type, rate_percent, description
		FROM akabou_surcharges WHERE id = ?
	`, id).Scan(&surcharge.ID, &surcharge.SurchargeType, &surcharge.RatePercent, &surcharge.Description)
	if err != nil {
		return nil, err
	}
	return surcharge, nil
}

// GetAllSurcharges 全割増料金を取得する
func (r *AkabouFareRepository) GetAllSurcharges() ([]*model.AkabouSurcharge, error) {
	rows, err := r.db.Query(`
		SELECT id, surcharge_type, rate_percent, description
		FROM akabou_surcharges ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var surcharges []*model.AkabouSurcharge
	for rows.Next() {
		s := &model.AkabouSurcharge{}
		if err := rows.Scan(&s.ID, &s.SurchargeType, &s.RatePercent, &s.Description); err != nil {
			return nil, err
		}
		surcharges = append(surcharges, s)
	}
	return surcharges, rows.Err()
}

// UpdateSurcharge 割増料金を更新する
func (r *AkabouFareRepository) UpdateSurcharge(surcharge *model.AkabouSurcharge) error {
	_, err := r.db.Exec(`
		UPDATE akabou_surcharges
		SET surcharge_type = ?, rate_percent = ?, description = ?
		WHERE id = ?
	`, surcharge.SurchargeType, surcharge.RatePercent, surcharge.Description, surcharge.ID)
	return err
}

// DeleteSurcharge 割増料金を削除する
func (r *AkabouFareRepository) DeleteSurcharge(id int64) error {
	_, err := r.db.Exec(`DELETE FROM akabou_surcharges WHERE id = ?`, id)
	return err
}

// === AkabouAreaSurcharge (地区割増) ===

// CreateAreaSurcharge 地区割増を作成する
func (r *AkabouFareRepository) CreateAreaSurcharge(area *model.AkabouAreaSurcharge) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO akabou_area_surcharges (area_name, surcharge_amount)
		VALUES (?, ?)
	`, area.AreaName, area.SurchargeAmount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAreaSurchargeByID IDで地区割増を取得する
func (r *AkabouFareRepository) GetAreaSurchargeByID(id int64) (*model.AkabouAreaSurcharge, error) {
	area := &model.AkabouAreaSurcharge{}
	err := r.db.QueryRow(`
		SELECT id, area_name, surcharge_amount
		FROM akabou_area_surcharges WHERE id = ?
	`, id).Scan(&area.ID, &area.AreaName, &area.SurchargeAmount)
	if err != nil {
		return nil, err
	}
	return area, nil
}

// GetAllAreaSurcharges 全地区割増を取得する
func (r *AkabouFareRepository) GetAllAreaSurcharges() ([]*model.AkabouAreaSurcharge, error) {
	rows, err := r.db.Query(`
		SELECT id, area_name, surcharge_amount
		FROM akabou_area_surcharges ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*model.AkabouAreaSurcharge
	for rows.Next() {
		a := &model.AkabouAreaSurcharge{}
		if err := rows.Scan(&a.ID, &a.AreaName, &a.SurchargeAmount); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, rows.Err()
}

// UpdateAreaSurcharge 地区割増を更新する
func (r *AkabouFareRepository) UpdateAreaSurcharge(area *model.AkabouAreaSurcharge) error {
	_, err := r.db.Exec(`
		UPDATE akabou_area_surcharges
		SET area_name = ?, surcharge_amount = ?
		WHERE id = ?
	`, area.AreaName, area.SurchargeAmount, area.ID)
	return err
}

// DeleteAreaSurcharge 地区割増を削除する
func (r *AkabouFareRepository) DeleteAreaSurcharge(id int64) error {
	_, err := r.db.Exec(`DELETE FROM akabou_area_surcharges WHERE id = ?`, id)
	return err
}

// === AkabouAdditionalFee (付帯料金) ===

// CreateAdditionalFee 付帯料金を作成する
func (r *AkabouFareRepository) CreateAdditionalFee(fee *model.AkabouAdditionalFee) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO akabou_additional_fees (fee_type, free_minutes, unit_minutes, fee_amount)
		VALUES (?, ?, ?, ?)
	`, fee.FeeType, fee.FreeMinutes, fee.UnitMinutes, fee.FeeAmount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAdditionalFeeByID IDで付帯料金を取得する
func (r *AkabouFareRepository) GetAdditionalFeeByID(id int64) (*model.AkabouAdditionalFee, error) {
	fee := &model.AkabouAdditionalFee{}
	err := r.db.QueryRow(`
		SELECT id, fee_type, free_minutes, unit_minutes, fee_amount
		FROM akabou_additional_fees WHERE id = ?
	`, id).Scan(&fee.ID, &fee.FeeType, &fee.FreeMinutes, &fee.UnitMinutes, &fee.FeeAmount)
	if err != nil {
		return nil, err
	}
	return fee, nil
}

// GetAllAdditionalFees 全付帯料金を取得する
func (r *AkabouFareRepository) GetAllAdditionalFees() ([]*model.AkabouAdditionalFee, error) {
	rows, err := r.db.Query(`
		SELECT id, fee_type, free_minutes, unit_minutes, fee_amount
		FROM akabou_additional_fees ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fees []*model.AkabouAdditionalFee
	for rows.Next() {
		f := &model.AkabouAdditionalFee{}
		if err := rows.Scan(&f.ID, &f.FeeType, &f.FreeMinutes, &f.UnitMinutes, &f.FeeAmount); err != nil {
			return nil, err
		}
		fees = append(fees, f)
	}
	return fees, rows.Err()
}

// UpdateAdditionalFee 付帯料金を更新する
func (r *AkabouFareRepository) UpdateAdditionalFee(fee *model.AkabouAdditionalFee) error {
	_, err := r.db.Exec(`
		UPDATE akabou_additional_fees
		SET fee_type = ?, free_minutes = ?, unit_minutes = ?, fee_amount = ?
		WHERE id = ?
	`, fee.FeeType, fee.FreeMinutes, fee.UnitMinutes, fee.FeeAmount, fee.ID)
	return err
}

// DeleteAdditionalFee 付帯料金を削除する
func (r *AkabouFareRepository) DeleteAdditionalFee(id int64) error {
	_, err := r.db.Exec(`DELETE FROM akabou_additional_fees WHERE id = ?`, id)
	return err
}
