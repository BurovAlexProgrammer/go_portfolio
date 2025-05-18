package configLoader

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	devPath  = "/config/dev.yaml"
	prodPath = "/config/prod.yaml"
)

type AppConfig struct {
	MysqlPath string     `yaml:"mysql_db_path" env-default:"/db/mysql.db"`
	HttpSrv   HttpServer `yaml:"http_server"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func New() *AppConfig {
	env := os.Getenv("ENV")
	if env == "" {
		slog.Warn("ENV not defined. ENV set as [dev] by default")
		env = "dev"
	}

	path := getConfigFullPath(devPath)
	if env == "prod" {
		path = getConfigFullPath(prodPath)
	}

	conf := AppConfig{}
	err := cleanenv.ReadConfig(path, &conf)

	if os.IsNotExist(err) {
		log.Fatal("Configuration file is not exist")
	}
	if err != nil {
		log.Fatal("Unexpected error while loading configuration file: \n", err)
	}

	return &conf
}

func getConfigFullPath(relativePath string) string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	return filepath.Join(dir, relativePath)
}
