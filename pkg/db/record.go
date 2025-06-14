package db

import (
	"database/sql"
	"fmt"

	"github.com/MRsummer/ChangeHairStyle/pkg/model"
)

// SaveHairStyleRecord 保存发型生成记录
func SaveHairStyleRecord(db interface{}, record *model.HairStyleRecord) error {
	query := `
		INSERT INTO hair_style_records (user_id, image_url, prompt)
		VALUES (?, ?, ?)
	`

	var result sql.Result
	var err error

	switch tx := db.(type) {
	case *sql.DB:
		result, err = tx.Exec(query, record.UserID, record.ImageURL, record.Prompt)
	case *sql.Tx:
		result, err = tx.Exec(query, record.UserID, record.ImageURL, record.Prompt)
	default:
		return fmt.Errorf("不支持的数据库连接类型")
	}

	if err != nil {
		return fmt.Errorf("保存生成记录失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取记录ID失败: %v", err)
	}

	record.ID = id
	return nil
}

// GetHairStyleRecords 获取用户的发型生成记录
func GetHairStyleRecords(db *sql.DB, userID string, page, pageSize int) (*model.RecordResponse, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取总记录数
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM hair_style_records
		WHERE user_id = ?
	`
	err := db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取记录总数失败: %v", err)
	}

	// 获取分页记录
	query := `
		SELECT id, user_id, image_url, prompt, created_at
		FROM hair_style_records
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("查询记录失败: %v", err)
	}
	defer rows.Close()

	// 解析记录
	var records []model.HairStyleRecord
	for rows.Next() {
		var record model.HairStyleRecord
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.ImageURL,
			&record.Prompt,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("解析记录失败: %v", err)
		}
		records = append(records, record)
	}

	return &model.RecordResponse{
		Total:   total,
		Records: records,
	}, nil
}
