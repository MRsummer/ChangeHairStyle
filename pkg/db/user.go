package db

import (
	"database/sql"
	"fmt"

	"github.com/MRsummer/ChangeHairStyle/pkg/model"
)

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(db *sql.DB, userInfo *model.UserInfo) error {
	query := `
        INSERT INTO user_info (user_id, nickname, avatar_url)
        VALUES (?, ?, ?)
        ON DUPLICATE KEY UPDATE
        nickname = VALUES(nickname),
        avatar_url = VALUES(avatar_url)
    `

	_, err := db.Exec(query, userInfo.UserID, userInfo.Nickname, userInfo.AvatarURL)
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}

	return nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(db *sql.DB, userID string) (*model.UserInfo, error) {
	query := `
        SELECT id, user_id, nickname, avatar_url, created_at, updated_at
        FROM user_info
        WHERE user_id = ?
    `

	userInfo := &model.UserInfo{}
	err := db.QueryRow(query, userID).Scan(
		&userInfo.ID,
		&userInfo.UserID,
		&userInfo.Nickname,
		&userInfo.AvatarURL,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	return userInfo, nil
}
