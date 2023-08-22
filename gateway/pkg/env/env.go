package env

import "os"

func GetWithDefault(variable, def string) string {
	if value, ok := os.LookupEnv(variable); ok {
		return value
	}
	return def
}
