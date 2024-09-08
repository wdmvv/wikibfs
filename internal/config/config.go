package config

import (
	"encoding/json"
	"io"
	"os"
)

type Cnf struct {
	Depth      int      `json:"depth"`
	Bases      []string `json:"bases"`
	Goroutines int      `json:"goroutines"`
	DbPath     string   `json:"dbpath"`
}

var Config Cnf

func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &Config)
}
