package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int // seconds
	MaxIdleTime  int // seconds
}

type CORSConfig struct {
	Origins []string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=" + d.SSLMode
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	expireHour, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	maxLifetime, _ := strconv.Atoi(getEnv("DB_MAX_LIFETIME_SEC", "300"))
	maxIdleTime, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_SEC", "180"))

	jwtSecret := getEnv("JWT_SECRET", "")
	mode := getEnv("GIN_MODE", "debug")

	// Reject weak/default JWT secrets in production
	if mode == "release" && (jwtSecret == "" || jwtSecret == "change-me-in-production") {
		log.Fatal("[安全] 生产环境必须设置 JWT_SECRET 环境变量（至少 32 位随机字符串）")
	}
	if jwtSecret == "" {
		jwtSecret = "dev-only-secret-not-for-production"
		log.Println("[警告] 使用开发环境 JWT 密钥，生产环境请设置 JWT_SECRET")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "1365"),
			Mode: mode,
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			DBName:       getEnv("DB_NAME", "ops_platform"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxLifetime:  maxLifetime,
			MaxIdleTime:  maxIdleTime,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			ExpireHour: expireHour,
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "ops-platform"),
			UseSSL:    useSSL,
		},
		CORS: CORSConfig{
			Origins: []string{
				getEnv("CORS_ORIGIN_1", "http://localhost:3000"),
				getEnv("CORS_ORIGIN_2", "http://localhost:1365"),
			},
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}
