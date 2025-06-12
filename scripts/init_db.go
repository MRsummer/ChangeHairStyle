package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 从环境变量获取数据库连接信息
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || user == "" || password == "" || dbname == "" {
		log.Fatal("请设置环境变量: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME")
	}

	if port == "" {
		port = "3306" // 默认端口
	}

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname,
	)

	// 连接数据库
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("测试数据库连接失败: %v", err)
	}

	fmt.Println("成功连接到数据库")

	// 读取SQL文件
	sqlBytes, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("读取SQL文件失败: %v", err)
	}

	// 将SQL文件内容分割为单独的语句
	sqlContent := string(sqlBytes)
	statements := strings.Split(sqlContent, ";")

	// 执行每个SQL语句
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err = db.Exec(statement)
		if err != nil {
			log.Fatalf("执行SQL失败: %v\nSQL: %s", err, statement)
		}
	}

	fmt.Println("成功执行SQL文件")
} 