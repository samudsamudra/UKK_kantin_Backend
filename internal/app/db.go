package app

import (
	"log"
	"os"

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

	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
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
		&Transaksi{},
		&DetailTransaksi{},
		&WalletTransaction{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations completed")
}
