package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config menampung semua variabel konfigurasi aplikasi.
// Menggunakan struct membuat konfigurasi menjadi eksplisit dan mudah dikelola.
type Config struct {
	APIVersion string
	Port       string
	JWTSecret  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// LoadConfig memuat konfigurasi dari file .env dan/atau environment variables.
func LoadConfig() *Config {
	// Memuat file .env, tidak akan error jika file tidak ada.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading, using environment variables.")
	}

	cfg := &Config{
		APIVersion: getEnv("API_VERSION", "v1"),
		Port:       getEnv("PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "random_string"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "sso_user"),
		DBPassword: getEnv("DB_PASSWORD", "sso_password"),
		DBName:     getEnv("DB_NAME", "sso_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set. Please set it in .env or system environment.")
	}

	return cfg
}

// getEnv adalah helper untuk mengambil env var dengan nilai default.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
