package config

import "os"

const (
	Source = "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable"
	Driver = "postgres"
)

type Config struct {
	Environment      string
	ConnectionString string
	Driver           string
}

func getConfigValue(envName string, defaultValue string) string {
	if val, ok := os.LookupEnv(envName); ok {
		return val
	}

	return defaultValue
}

func New() Config {
	return Config{
		Environment:      getConfigValue("ENV", "local"),
		ConnectionString: getConfigValue("CONN_STRING", Source),
		Driver:           getConfigValue("CONN_DRIVER", Driver),
	}
}
