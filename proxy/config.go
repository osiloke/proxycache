package proxy

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Target      string `json:"target"`
	Debug       bool   `json:"debug"`
	CacheFolder string `json:"cache_folder"`
	Port        string `json:"port"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config
	json.Unmarshal(file, &config)

	return &config, nil
}
