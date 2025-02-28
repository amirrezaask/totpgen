package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kirsle/configdir"
	"github.com/pquerna/otp/totp"
)

type Account struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}
type Config []Account

const CONFIG_SAMPLE = `[{"name": "Sample", "secret": "oops"}]`

func createConfigSampleFile() error {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "2fa.json")

	err := os.WriteFile(configPath, []byte(CONFIG_SAMPLE), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func readConfig(path string) (Config, error) {
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

	return cfg, nil
}

func generateCmd() error {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "2fa.json")
	cfg, err := readConfig(configPath)
	if err != nil {
		return err
	}
	t := time.Now()
	for _, acc := range cfg {
		code, err := totp.GenerateCode(acc.Secret, t)
		if err != nil {
			return err
		}
		fmt.Printf("%s => %s\n", acc.Name, code)
	}

	return nil
}

func generateSingleCmd(name string) error {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "2fa.json")
	cfg, err := readConfig(configPath)
	if err != nil {
		return err
	}
	t := time.Now()
	for _, acc := range cfg {
		re := regexp.MustCompile(name)
		if re.Match([]byte(acc.Name)) {
			code, err := totp.GenerateCode(acc.Secret, t)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", code)
		}
	}

	return nil
}

func getEditorExecCommand() *exec.Cmd {
	EDITOR := os.Getenv("EDITOR")
	VISUAL := os.Getenv("VISUAL")
	if EDITOR != "" {
		segs := strings.Split(EDITOR, " ")
		editorProgram := segs[0]
		var args []string
		if len(segs) > 1 {
			args = segs[1:]
		}
		return exec.Command(editorProgram, args...)
	} else if VISUAL != "" {
		segs := strings.Split(VISUAL, " ")
		editorProgram := segs[0]
		var args []string
		if len(segs) > 1 {
			args = segs[1:]
		}
		return exec.Command(editorProgram, args...)
	} else if runtime.GOOS == "windows" {
		_, err := exec.LookPath("code")
		if err != nil {
			return exec.Command("notepad")
		}
		return exec.Command("code", "-w")
	} else {
		return exec.Command("nano")
	}
}

func editCmd() error {
	path := configdir.LocalConfig()
	configPath := filepath.Join(path, "2fa.json")
	editor := getEditorExecCommand()
	editor.Args = append(editor.Args, configPath)

	if err := editor.Run(); err != nil {
		return err
	}
	return nil
}

// TODO:
// Edit command to edit configuration in system editor
func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		if err := generateCmd(); err != nil {
			panic(fmt.Sprintf("running generate cmd failed: %s", err.Error()))
		}
	} else {
		cmd := args[0]
		switch cmd {
		case "edit":
			if err := editCmd(); err != nil {
				panic(fmt.Sprintf("running edit cmd failed: %s", err.Error()))
			}
		default:
			// case you want just one of the passwords
			if err := generateSingleCmd(cmd); err != nil {
				panic(fmt.Sprintf("running single generate failed: %s", err.Error()))
			}
		}
	}
}
