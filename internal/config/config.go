package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

const (
	EnvDev            = "dev"
	EnvTest           = "test"
	EnvProd           = "prod"
	FileUploadS3Place = "s3"
)

type Config struct {
	Env               string `env:"ENV" env-required:"true"`
	BlockingParanoia  int    `env:"BLOCKING_PARANOIA" env-default:"2"`
	ThisServiceDomain string `env:"THIS_SERVICE_DOMAIN" env-required:"true"`
	UploadPlace       string `env:"FILE_UPLOAD_PLACE" env-required:"true"`
	HTTP              HTTPServer
	DB                Database
	Cors              Cors
	File              File
	Drive             Drive
	S3                S3
	RateLimiter       RateLimiter
}

type HTTPServer struct {
	Host         string        `env:"HTTP_HOST" env-default:"localhost"`
	Port         uint16        `env:"HTTP_PORT" env-default:"8083"`
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

type RateLimiter struct {
	AllowanceRequests int `env:"RATE_LIMITER_ALLOWANCE_REQUESTS" env-default:"150"`
	TimeDurationSec   int `env:"RATE_LIMITER_TIME_DURATION_SECONDS" env-default:"180"`
}

type File struct {
	UploadMaxSize       int64  `env:"FILE_UPLOAD_MAX_SIZE" env-required:"true"`
	LimitStoragePerUser int64  `env:"FILE_LIMIT_STORAGE_PER_USER" env-required:"true"`
	SavePath            string `env:"FILE_SAVE_PATH" env-default:"./uploads/user_files"`
}

type Drive struct {
	UploadMaxSize int64  `env:"DRIVE_UPLOAD_MAX_SIZE"`
	LimitPerUser  int64  `env:"DRIVE_LIMIT_STORAGE_PER_USER"`
	SavePath      string `env:"FILE_SAVE_PATH" env-default:"./uploads/drive"`
	UseEncryption bool   `env:"DRIVE_USE_FILE_ENCRYPTION" env-default:"false"`
	EncryptionKey string `env:"DRIVE_ENCRYPTION_KEY" env-default:""`
}

type S3 struct {
	Endpoint        string `env:"S3_ENDPOINT" env-default:""`
	AccessKey       string `env:"S3_ACCESS_KEY" env-default:""`
	SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY" env-default:""`
	UseSSL          bool   `env:"S3_USE_SSL" env-default:"false"`
	BucketName      string `env:"S3_BUCKET_NAME" env-default:""`
	Location        string `env:"S3_LOCATION" env-default:""`
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
