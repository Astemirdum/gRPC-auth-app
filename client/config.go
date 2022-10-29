package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var cfg *Config

type Config struct {
	Service Service `yaml:"service"`
}

type Service struct {
	Addr string `yaml:"addr"`
}

func ReadConfigYML(configYML string) *Config {
	if cfg != nil {
		return cfg
	}
	file, err := os.Open(configYML)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	jscfg, _ := json.MarshalIndent(cfg, "", "	")
	fmt.Println(string(jscfg))

	return cfg
}
