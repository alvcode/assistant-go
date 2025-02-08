package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

type Config struct {
	Env  string `env:"ENV" env-required:"true"`
	HTTP HTTPServer
	DB   Database
	Cors Cors
}

type HTTPServer struct {
	Host         string        `env:"HTTP_HOST" env-default:"localhost"`
	Port         uint16        `env:"HTTP_PORT" env-default:"8083"`
	Timeout      time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"15s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"15s"`
}

type Database struct {
	Driver   string `env:"DB_DRIVER" env-required:"true"`
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     string `env:"DB_PORT" env-required:"true"`
	Username string `env:"DB_USERNAME" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Database string `env:"DB_DATABASE" env-required:"true"`
}

type Cors struct {
	AllowedMethods     string `env:"CORS_ALLOWED_METHODS" env-required:"true"`
	AllowedOrigins     string `env:"CORS_ALLOWED_ORIGINS" env-required:"true"`
	AllowedHeaders     string `env:"CORS_ALLOWED_HEADERS" env-required:"true"`
	AllowCredentials   bool   `env:"CORS_ALLOW_CREDENTIALS" env-required:"true"`
	OptionsPassthrough bool   `env:"CORS_OPTIONS_PASSTHROUGH" env-required:"true"`
	ExposedHeaders     string `env:"CORS_EXPOSED_HEADERS" env-required:"true"`
	Debug              bool   `env:"CORS_DEBUG" env-default:"false"`
}

const configFilePath = ".env"

func MustLoad() *Config {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err == nil {
		log.Println("config variables loaded")
		return &cfg
	}

	log.Printf("error read config variables: %s ", err)

	log.Println("Trying to load from a .env file")
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Fatal("config file .env does not exist")
	}

	err = cleanenv.ReadConfig(configFilePath, &cfg)
	if err != nil {
		log.Fatalf("error reading .env file: %s", err)
	}
	log.Println("config .env file loaded")
	return &cfg
}
