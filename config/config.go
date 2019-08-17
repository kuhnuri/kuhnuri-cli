package config

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Config map[string]string

var filenames []string

func init() {
	file := ".kuhnurirc"
	filenames = []string{}
	dir, err := filepath.Abs(".")
	if err != nil {
		panic(fmt.Sprintf("Failed to get absolute path: %s", err))
	}
	for {
		filenames = append(filenames, filepath.Join(dir, file))
		parent := filepath.Dir(dir)
		if dir == parent {
			break
		}
		dir = parent
	}
	usr, err := user.Current()
	if err != nil {
		// Ignore
	} else {
		filenames = append(filenames, filepath.Join(usr.HomeDir, file))
	}
}

func Read() Config {
	config := Config{}
	for _, filename := range filenames {
		abs, err := filepath.Abs(filename)
		if err != nil {
			continue
		}
		_, err = os.Stat(abs)
		if os.IsNotExist(err) {
			continue
		}
		file, err := os.Open(abs)
		if err != nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "#") {
				continue
			}
			if equal := strings.Index(line, "="); equal >= 0 {
				if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
					value := ""
					if len(line) > equal {
						value = strings.TrimSpace(line[equal+1:])
					}
					config[key] = value
				}
			}
		}
		if err := scanner.Err(); err != nil {
			continue
		}
	}
	return config
}
