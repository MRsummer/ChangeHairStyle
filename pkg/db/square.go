package db

import (
	"database/sql"
	"fmt"

	"github.com/MRsummer/ChangeHairStyle/pkg/model"
)

// ShareToSquare 分享到广场
func ShareToSquare(db *sql.DB, content *model.SquareContent) error {
	query := `
        INSERT INTO square_content (user_id, record_id)
        VALUES (?, ?)
    `

	result, err := db.Exec(query, content.UserID, content.RecordID)
	if err != nil {
		return fmt.Errorf("分享到广场失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}

	content.ID = id
	return nil
}

// GetSquareContents 获取广场内容列表
func GetSquareContents(db *sql.DB, userID string, page, pageSize int) (*model.SquareContentResponse, error) {
	// 获取总记录数
	countQuery := `SELECT COUNT(*) FROM square_content`
	var total int64
	err := db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取总记录数失败: %v", err)
	}

	// 获取分页数据
	query := `
        SELECT 
            sc.id, sc.user_id, sc.record_id, sc.like_count, sc.created_at, sc.updated_at,
            hr.image_url, hr.prompt, hr.created_at as record_created_at,
            ui.nickname, ui.avatar_url,
            CASE WHEN lr.id IS NOT NULL THEN 1 ELSE 0 END as is_liked
        FROM square_content sc
        LEFT JOIN hair_style_record hr ON sc.record_id = hr.id
        LEFT JOIN user_info ui ON sc.user_id = ui.user_id
        LEFT JOIN like_record lr ON sc.id = lr.content_id AND lr.user_id = ?
        ORDER BY sc.created_at DESC
        LIMIT ? OFFSET ?
    `

	offset := (page - 1) * pageSize
	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("查询广场内容失败: %v", err)
	}
	defer rows.Close()

	var contents []model.SquareContent
	for rows.Next() {
		var content model.SquareContent
		var record model.HairStyleRecord
		var userInfo model.UserInfo

		err := rows.Scan(
			&content.ID,
			&content.UserID,
			&content.RecordID,
			&content.LikeCount,
			&content.CreatedAt,
			&content.UpdatedAt,
			&record.ImageURL,
			&record.Prompt,
			&record.CreatedAt,
			&userInfo.Nickname,
			&userInfo.AvatarURL,
			&content.IsLiked,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}

		content.Record = &record
		content.UserInfo = &userInfo
		contents = append(contents, content)
	}

	return &model.SquareContentResponse{
		Total:   total,
		Records: contents,
	}, nil
}

// LikeContent 点赞内容
func LikeContent(db *sql.DB, userID string, contentID int64) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查是否已点赞
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM like_record WHERE user_id = ? AND content_id = ?)",
		userID, contentID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查点赞状态失败: %v", err)
	}

	if exists {
		// 取消点赞
		_, err = tx.Exec("DELETE FROM like_record WHERE user_id = ? AND content_id = ?",
			userID, contentID)
		if err != nil {
			return fmt.Errorf("取消点赞失败: %v", err)
		}

		_, err = tx.Exec("UPDATE square_content SET like_count = like_count - 1 WHERE id = ?",
			contentID)
		if err != nil {
			return fmt.Errorf("更新点赞数失败: %v", err)
		}
	} else {
		// 添加点赞
		_, err = tx.Exec("INSERT INTO like_record (user_id, content_id) VALUES (?, ?)",
			userID, contentID)
		if err != nil {
			return fmt.Errorf("添加点赞失败: %v", err)
		}

		_, err = tx.Exec("UPDATE square_content SET like_count = like_count + 1 WHERE id = ?",
			contentID)
		if err != nil {
			return fmt.Errorf("更新点赞数失败: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
