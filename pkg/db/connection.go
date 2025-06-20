package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/MRsummer/ChangeHairStyle/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

// InitDB 初始化数据库连接
func InitDB() (*sql.DB, error) {
	// 从环境变量获取数据库连接信息
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname,
	)

	// 连接数据库
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		logger.WithError(err).Error("打开数据库连接失败")
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := db.Ping(); err != nil {
		logger.WithError(err).Error("数据库连接测试失败")
		return nil, fmt.Errorf("测试数据库连接失败: %v", err)
	}

	return db, nil
}
