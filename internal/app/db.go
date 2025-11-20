package app

import (
	"database/sql"
	"log"
	"os"

	goMySQL "github.com/go-sql-driver/mysql"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	// try connect directly
	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err == nil {
		DB = db
		log.Println("database connected")
		return
	}

	// if failed, attempt to create database if unknown
	// parse DSN using go-sql-driver/mysql to extract DB name
	cfg, perr := goMySQL.ParseDSN(dsn)
	if perr != nil {
		log.Fatalf("invalid DSN: %v", perr)
	}
	dbName := cfg.DBName
	if dbName == "" {
		log.Fatalf("DSN has no database name: %s", dsn)
	}

	// connect without DB name
	cfg.DBName = ""
	dsnNoDB := cfg.FormatDSN()
	sqlDB, err := sql.Open("mysql", dsnNoDB)
	if err != nil {
		log.Fatalf("failed to open raw sql connection: %v", err)
	}
	defer sqlDB.Close()

	createSQL := "CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
	if _, err := sqlDB.Exec(createSQL); err != nil {
		log.Fatalf("failed to create database %s: %v", dbName, err)
	}
	log.Printf("database %s ensured", dbName)

	// try connect again with gorm
	db, err = gorm.Open(gormMySQL.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database after creating: %v", err)
	}
	DB = db
	log.Println("database connected")
}

func RunMigrations() {
	if DB == nil {
		log.Fatal("DB is nil; InitDB must be called first")
	}
	err := DB.AutoMigrate(
		&User{},
		&Siswa{},
		&Stan{},
		&Menu{},
		&Diskon{},
		&MenuDiskon{},
		&Transaksi{},
		&DetailTransaksi{},
		&MenuDiskon{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations completed")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_diskon_time ON diskons (tanggal_awal, tanggal_akhir)")
}

