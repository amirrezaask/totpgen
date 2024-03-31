package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kirsle/configdir"
	"github.com/pquerna/otp/totp"
)

type Account struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}
type Config struct {
	Accounts []Account `json:"accounts"`
}

const CONFIG_SAMPLE = `{"accounts": [{"name": "Sample", "secret": "oops"}]}`

func createConfigSampleFile() error {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "totpgen.json")

	err := os.WriteFile(configPath, []byte(CONFIG_SAMPLE), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func readConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = createConfigSampleFile()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.NewDecoder(fd).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// TODO:
// Edit command to edit configuration in system editor
func main() {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "totpgen.json")
	cfg, err := readConfig(configPath)
	if err != nil {
		panic(err)
	}
	t := time.Now()
	for _, acc := range cfg.Accounts {
		code, err := totp.GenerateCode(acc.Secret, t)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s :: %s\n", acc.Name, code)
	}
}
