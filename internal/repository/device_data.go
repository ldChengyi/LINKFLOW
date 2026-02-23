package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
)

type DeviceDataRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceDataRepository(pool *pgxpool.Pool) *DeviceDataRepository {
	return &DeviceDataRepository{pool: pool}
}

func (r *DeviceDataRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

// InsertTelemetry 插入遥测数据
func (r *DeviceDataRepository) InsertTelemetry(ctx context.Context, deviceID, userID, topic string, payload map[string]interface{}, qos byte, valid bool, errors map[string]string) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var errorsJSON []byte
	if len(errors) > 0 {
		errorsJSON, _ = json.Marshal(errors)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO device_data (time, device_id, user_id, topic, payload, qos, valid, errors)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, time.Now(), deviceID, userID, topic, payloadJSON, qos, valid, errorsJSON)
	return err
}

// DeviceLatestData 设备最新遥测数据
type DeviceLatestData struct {
	Time    time.Time              `json:"time"`
	Payload map[string]interface{} `json:"payload"`
	Valid   bool                   `json:"valid"`
	Errors  map[string]string      `json:"errors,omitempty"`
}

// DeviceHistoryData 历史遥测数据
type DeviceHistoryData struct {
	Time    time.Time              `json:"time"`
	Payload map[string]interface{} `json:"payload"`
	Valid   bool                   `json:"valid"`
}

// GetDataHistory 获取设备历史遥测数据
func (r *DeviceDataRepository) GetDataHistory(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]DeviceHistoryData, error) {
	rows, err := r.query(ctx, `
		SELECT time, payload, valid
		FROM device_data
		WHERE device_id = $1 AND time BETWEEN $2 AND $3
		ORDER BY time ASC
		LIMIT $4
	`, deviceID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DeviceHistoryData
	for rows.Next() {
		var d DeviceHistoryData
		var payloadRaw []byte
		if err := rows.Scan(&d.Time, &payloadRaw, &d.Valid); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payloadRaw, &d.Payload); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, nil
}

func (r *DeviceDataRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

// AggregatedDataPoint 聚合数据点（avg/max/min）
type AggregatedDataPoint struct {
	Time       time.Time          `json:"time"`
	Payload    map[string]float64 `json:"payload"`
	MaxPayload map[string]float64 `json:"max_payload"`
	MinPayload map[string]float64 `json:"min_payload"`
}

// GetDataHistoryAggregated 按时间窗口聚合查询历史遥测数据（avg/max/min）
func (r *DeviceDataRepository) GetDataHistoryAggregated(
	ctx context.Context, deviceID string,
	start, end time.Time, interval string,
) ([]AggregatedDataPoint, error) {
	rows, err := r.query(ctx, `
		SELECT
		  time_bucket($1::interval, time) AS bucket,
		  kv.key,
		  ROUND(AVG(kv.value::numeric), 2) AS avg_val,
		  ROUND(MAX(kv.value::numeric), 2) AS max_val,
		  ROUND(MIN(kv.value::numeric), 2) AS min_val
		FROM device_data,
		  LATERAL (
		    SELECT key, value
		    FROM jsonb_each_text(payload)
		    WHERE value ~ '^-?[0-9]+(\.[0-9]+)?$'
		  ) kv(key, value)
		WHERE device_id = $2 AND time BETWEEN $3 AND $4
		GROUP BY bucket, kv.key
		ORDER BY bucket ASC
	`, interval, deviceID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// pivot: bucket → AggregatedDataPoint
	bucketMap := make(map[time.Time]*AggregatedDataPoint)
	var bucketOrder []time.Time

	for rows.Next() {
		var bucket time.Time
		var key string
		var avgVal, maxVal, minVal float64
		if err := rows.Scan(&bucket, &key, &avgVal, &maxVal, &minVal); err != nil {
			return nil, err
		}
		if _, ok := bucketMap[bucket]; !ok {
			bucketMap[bucket] = &AggregatedDataPoint{
				Time:       bucket,
				Payload:    make(map[string]float64),
				MaxPayload: make(map[string]float64),
				MinPayload: make(map[string]float64),
			}
			bucketOrder = append(bucketOrder, bucket)
		}
		pt := bucketMap[bucket]
		pt.Payload[key] = avgVal
		pt.MaxPayload[key] = maxVal
		pt.MinPayload[key] = minVal
	}

	result := make([]AggregatedDataPoint, 0, len(bucketOrder))
	for _, b := range bucketOrder {
		result = append(result, *bucketMap[b])
	}
	return result, nil
}

// GetLatestData 获取设备最新一条遥测数据
func (r *DeviceDataRepository) GetLatestData(ctx context.Context, deviceID string) (*DeviceLatestData, error) {
	var (
		t          time.Time
		payloadRaw []byte
		valid      bool
		errorsRaw  []byte
	)

	err := r.queryRow(ctx, `
		SELECT time, payload, valid, errors
		FROM device_data
		WHERE device_id = $1
		ORDER BY time DESC
		LIMIT 1
	`, deviceID).Scan(&t, &payloadRaw, &valid, &errorsRaw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	data := &DeviceLatestData{Time: t, Valid: valid}
	if err := json.Unmarshal(payloadRaw, &data.Payload); err != nil {
		return nil, err
	}
	if errorsRaw != nil {
		json.Unmarshal(errorsRaw, &data.Errors)
	}
	return data, nil
}
