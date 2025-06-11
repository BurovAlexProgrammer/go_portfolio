package ORM

import (
	"GoPortfolio/internal/domain"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewPostgresDB(dsn string, debug bool) (*gorm.DB, error) {
	logLevel := logger.Warn
	if debug {
		logLevel = logger.Info
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // медленные запросы (опционально)
			LogLevel:                  logLevel,    // уровень логирования
			IgnoreRecordNotFoundError: true,        // не логировать gorm.ErrRecordNotFound
			Colorful:                  true,        // цветной вывод
		},
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
}

func Automigrate(db *gorm.DB) error {
	models := []interface{}{
		&domain.User{},
		&domain.Task{},
	}
	return db.AutoMigrate(models...)
}

func GetPostgresDSNFromEnv() string {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DBNAME")
	sslmode := os.Getenv("PG_SSLMODE")
	timezone := os.Getenv("PG_TIMEZONE")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, password, dbname, sslmode, timezone,
	)
}
