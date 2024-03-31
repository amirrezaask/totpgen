package main

import (
	"encoding/json"
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

func createConfigDirectoryAndFile() error {
	path := configdir.LocalConfig("otpgen")
	err := os.Mkdir(path, 0644)
	if err != nil && !os.IsExist(err) {
		return err
	}

	configPath := filepath.Join(path, "config.json")

	err = os.WriteFile(configPath, []byte(CONFIG_SAMPLE), os.FileMode(os.O_CREATE))
	if err != nil {
		return err
	}

	return nil
}

func readConfig(path string) (*Config, error) {
	fd, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		err = createConfigDirectoryAndFile()
		if err != nil {
			return nil, err
		}
		fd, err = os.Open(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	var cfg Config
	err = json.NewDecoder(fd).Decode(&cfg)
	if err != nil {
		fmt.Println("inja3")
		return nil, err
	}

	return &cfg, nil
}

func main() {
	path := configdir.LocalConfig("otpgen")
	configPath := filepath.Join(path, "config.json")
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
