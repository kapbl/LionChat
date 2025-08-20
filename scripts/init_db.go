package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SqlFile  string
}

func main() {
	config := &Config{}

	// 命令行参数
	flag.StringVar(&config.Host, "host", "bj-cynosdbmysql-grp-ruts707k.sql.tencentcdb.com", "MySQL主机地址")
	flag.StringVar(&config.Port, "port", "24612", "MySQL端口")
	flag.StringVar(&config.Username, "username", "root", "MySQL用户名")
	flag.StringVar(&config.Password, "password", "0220059cyCY", "MySQL密码")
	flag.StringVar(&config.Database, "database", "lionchat", "数据库名")
	flag.StringVar(&config.SqlFile, "file", "./chat.sql", "SQL文件路径")
	flag.Parse()

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("获取当前工作目录失败:", err)
	}

	// 如果SQL文件路径不是绝对路径，则相对于当前工作目录
	if !filepath.IsAbs(config.SqlFile) {
		// 如果当前目录不是scripts目录，尝试在scripts子目录中查找
		if filepath.Base(wd) != "scripts" {
			scriptsPath := filepath.Join(wd, "scripts", config.SqlFile)
			if _, err := os.Stat(scriptsPath); err == nil {
				config.SqlFile = scriptsPath
			} else {
				config.SqlFile = filepath.Join(wd, config.SqlFile)
			}
		} else {
			config.SqlFile = filepath.Join(wd, config.SqlFile)
		}
	}

	// 检查SQL文件是否存在
	if _, err := os.Stat(config.SqlFile); os.IsNotExist(err) {
		log.Fatalf("SQL文件不存在: %s", config.SqlFile)
	}

	// 如果密码为空，提示用户输入
	if config.Password == "" {
		fmt.Print("请输入MySQL密码: ")
		fmt.Scanln(&config.Password)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", config.Username, config.Password, config.Host, config.Port)

	fmt.Printf("连接数据库: %s@%s:%s/%s\n", config.Username, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 读取SQL文件
	fmt.Printf("读取SQL文件: %s\n", config.SqlFile)
	sqlContent, err := ioutil.ReadFile(config.SqlFile)
	if err != nil {
		log.Fatal("读取SQL文件失败:", err)
	}

	// 执行SQL
	fmt.Println("开始执行SQL语句...")
	if err := executeSQLFile(db, string(sqlContent)); err != nil {
		log.Fatal("执行SQL失败:", err)
	}

	fmt.Println("✅ 数据库初始化完成！")

	// 验证表是否创建成功
	fmt.Println("验证表创建情况...")
	verifyTables(db)
}

// 执行SQL文件中的多个语句
func executeSQLFile(db *sql.DB, sqlContent string) error {
	// 按分号分割SQL语句
	statements := strings.Split(sqlContent, ";")

	for i, stmt := range statements {
		// 清理语句（去除空白字符和注释）
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// 跳过注释行
		if strings.HasPrefix(stmt, "--") || strings.HasPrefix(stmt, "/*") {
			continue
		}

		fmt.Printf("执行语句 %d/%d...\n", i+1, len(statements))

		// 执行语句
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("执行语句失败: %v\n语句: %s", err, stmt[:min(len(stmt), 100)])
		}
	}

	return nil
}

// 获取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 验证表是否创建成功
func verifyTables(db *sql.DB) {
	tables := []string{
		"users", "message", "group", "group_member",
		"user_friends", "moment", "comment", "like", "timeline",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = '%s'", table)).Scan(&count)
		if err != nil {
			fmt.Printf("❌ 检查表 %s 失败: %v\n", table, err)
			continue
		}

		if count > 0 {
			fmt.Printf("✅ 表 %s 创建成功\n", table)
		} else {
			fmt.Printf("❌ 表 %s 未找到\n", table)
		}
	}
}
