package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	WriteDBURL    string
	ReadDBURL     string
	EnableReplica bool
	APIKey        string
	LogFile       string
}

func Load() Config {
	loadDotEnv(".env")
	addr := getEnv("HTTP_ADDR", ":8080")
	writeDB := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/smart_inventory?sslmode=disable")
	readDB := getEnv("READ_DATABASE_URL", "")
	apiKey := getEnv("API_KEY", "")
	logFile := getEnv("LOG_FILE", "")
	return Config{
		HTTPAddr:      addr,
		WriteDBURL:    writeDB,
		ReadDBURL:     readDB,
		EnableReplica: readDB != "",
		APIKey:        apiKey,
		LogFile:       logFile,
	}
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
