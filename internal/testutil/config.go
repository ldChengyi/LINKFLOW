package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DatabaseConfig struct {
	Host          string `json:"host"`
	Port          string `json:"port"`
	Name          string `json:"name"`
	AdminUser     string `json:"admin_user"`
	AdminPassword string `json:"admin_password"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type JWTConfig struct {
	Secret      string `json:"secret"`
	ExpireHours int    `json:"expire_hours"`
}

type IntegrationConfig struct {
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	JWT      JWTConfig      `json:"jwt"`
}

type TestConfig struct {
	Accounts     map[string]Account `json:"accounts"`
	InvalidCases map[string]Account `json:"invalid_cases"`
	Integration  IntegrationConfig  `json:"integration"`
}

func LoadTestConfig() (*TestConfig, error) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "testdata", "auth_test.json"))
	if err != nil {
		return nil, err
	}
	var cfg TestConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
