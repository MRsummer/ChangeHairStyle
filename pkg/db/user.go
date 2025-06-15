package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/MRsummer/ChangeHairStyle/pkg/model"
)

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(db *sql.DB, userInfo *model.UserInfo) error {
	query := `
        UPDATE user_info 
        SET nickname = ?, avatar_url = ?
        WHERE user_id = ?
    `

	result, err := db.Exec(query, userInfo.Nickname, userInfo.AvatarURL, userInfo.UserID)
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		fmt.Printf("[UpdateUserInfo] 用户不存在: userID=%s\n", userInfo.UserID)
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(db *sql.DB, userID string) (*model.UserInfo, error) {
	query := `
        SELECT id, user_id, nickname, avatar_url, coin, created_at, updated_at
        FROM user_info
        WHERE user_id = ?
    `

	userInfo := &model.UserInfo{}
	var nickname, avatarURL sql.NullString
	err := db.QueryRow(query, userID).Scan(
		&userInfo.ID,
		&userInfo.UserID,
		&nickname,
		&avatarURL,
		&userInfo.Coin,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 将 sql.NullString 转换为 string
	if nickname.Valid {
		userInfo.Nickname = nickname.String
	}
	if avatarURL.Valid {
		userInfo.AvatarURL = avatarURL.String
	}

	return userInfo, nil
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode(db *sql.DB, userID string) (string, error) {
	// 先查询是否已有邀请码
	var existingCode string
	err := db.QueryRow("SELECT invite_code FROM user_info WHERE user_id = ?", userID).Scan(&existingCode)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("查询邀请码失败: %v", err)
	}

	// 如果已有邀请码，直接返回
	if existingCode != "" {
		return existingCode, nil
	}

	// 生成6位随机邀请码
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	inviteCode := string(b)

	// 更新用户邀请码
	query := `UPDATE user_info SET invite_code = ? WHERE user_id = ?`
	result, err := db.Exec(query, inviteCode, userID)
	if err != nil {
		return "", fmt.Errorf("更新邀请码失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return "", fmt.Errorf("用户不存在")
	}

	return inviteCode, nil
}

// UseInviteCode 使用邀请码
func UseInviteCode(db *sql.DB, userID, inviteCode string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查是否已使用过邀请码
	var usedInviteCode string
	err = tx.QueryRow("SELECT used_invite_code FROM user_info WHERE user_id = ?", userID).Scan(&usedInviteCode)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查邀请码使用状态失败: %v", err)
	}
	if usedInviteCode != "" {
		return fmt.Errorf("您已使用过邀请码")
	}

	// 查找邀请人并增加coin
	result, err := tx.Exec("UPDATE user_info SET coin = coin + 20 WHERE invite_code = ?", inviteCode)
	if err != nil {
		return fmt.Errorf("更新邀请人coin失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("邀请码无效")
	}

	// 标记用户已使用邀请码
	_, err = tx.Exec("UPDATE user_info SET used_invite_code = ? WHERE user_id = ?", inviteCode, userID)
	if err != nil {
		return fmt.Errorf("更新用户邀请码使用状态失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// SignIn 用户签到
func SignIn(db *sql.DB, userID string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查今日是否已签到
	var lastSignInDate sql.NullTime
	err = tx.QueryRow("SELECT last_sign_in_date FROM user_info WHERE user_id = ?", userID).Scan(&lastSignInDate)
	if err != nil {
		return fmt.Errorf("查询签到记录失败: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if lastSignInDate.Valid && lastSignInDate.Time.Format("2006-01-02") == today {
		return fmt.Errorf("今日已签到")
	}

	// 更新签到时间和coin
	_, err = tx.Exec("UPDATE user_info SET last_sign_in_date = CURDATE(), coin = coin + 5 WHERE user_id = ?",
		userID)
	if err != nil {
		return fmt.Errorf("更新签到信息失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// CheckCoin 检查用户金币是否足够
func CheckCoin(db *sql.DB, userID string, amount int) (bool, error) {
	var coin int
	err := db.QueryRow("SELECT coin FROM user_info WHERE user_id = ?", userID).Scan(&coin)
	if err != nil {
		return false, fmt.Errorf("查询coin失败: %v", err)
	}
	return coin >= amount, nil
}

// DeductCoin 扣除用户金币
func DeductCoin(db *sql.DB, userID string, amount int) error {
	result, err := db.Exec("UPDATE user_info SET coin = coin - ? WHERE user_id = ?", amount, userID)
	if err != nil {
		return fmt.Errorf("扣除coin失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB, userInfo *model.UserInfo) error {
	query := `
        INSERT INTO user_info (user_id, coin)
        VALUES (?, ?)
    `

	result, err := db.Exec(query, userInfo.UserID, userInfo.Coin)
	if err != nil {
		return fmt.Errorf("创建用户失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取用户ID失败: %v", err)
	}

	userInfo.ID = id
	return nil
}
