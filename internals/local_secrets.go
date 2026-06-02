package internals

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type LocalSecrets struct{}

// Load implements SecretsInterface.
func (LocalSecrets) Load() error {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &configNotFound) {
			log.Println(".env not found, continuing with OS environment")
			return nil
		}

		log.Println("local secrets: unable to parse .env, continuing with OS environment:", err)
	}

	return nil
}

// Get implements SecretsInterface.
func (LocalSecrets) Get(key string) string {
	value := viper.GetString(key)
	if value != "" {
		_ = os.Setenv(key, value)
		return value
	}

	value = os.Getenv(key)
	if value != "" {
		return value
	}

	value = readSimpleEnvValue(".env", key)
	if value != "" {
		_ = os.Setenv(key, value)
	}

	return value
}

// GetOrPanic implements SecretsInterface.
func (l LocalSecrets) GetOrPanic(key string) string {
	value := l.Get(key)
	if value == "" {
		log.Panicf("%s is empty", key)
	}

	return value
}

func (*LocalSecrets) NewLocalSecret() ISecrets {
	return LocalSecrets{}
}

// readSimpleEnvValue reads KEY=VALUE entries from .env, including multiline triple-quoted values.
func readSimpleEnvValue(filePath, targetKey string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var multilineKey string
	var multilineBuilder strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if multilineKey != "" {
			if line == `"""` {
				if multilineKey == targetKey {
					return multilineBuilder.String()
				}
				multilineKey = ""
				multilineBuilder.Reset()
				continue
			}

			if multilineBuilder.Len() > 0 {
				multilineBuilder.WriteString("\n")
			}
			multilineBuilder.WriteString(line)
			continue
		}

		if !strings.Contains(line, "=") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if value == `"""` {
			multilineKey = key
			multilineBuilder.Reset()
			continue
		}

		if key == targetKey {
			return strings.Trim(value, `"'`)
		}
	}

	return ""
}
