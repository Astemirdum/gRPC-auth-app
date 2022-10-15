package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	once sync.Once
	cfg  *Config
)

type Grpc struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"user"`
	Password string `yaml:"password"`
	NameDB   string `yaml:"dbname"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

type Clickhouse struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	Username string `yaml:"username"`
	NameDB   string `yaml:"db"`
}

type Kafka struct {
	Addr          string `yaml:"addr"`
	Topic         string `yaml:"topic"`
	ConsumerGroup string `yaml:"consumer-group"`
	ConsumerNum   int    `yaml:"consumer-num"`
	AddrClick     string `yaml:"addr-click"`
	PartitionNum  int    `yaml:"partition-num"`
	Format        string `yaml:"format"`
}

type Config struct {
	Grpc       Grpc       `yaml:"grpc"`
	Database   DB         `yaml:"db"`
	Redis      Redis      `yaml:"redis"`
	Kafka      Kafka      `yaml:"kafka"`
	Clickhouse Clickhouse `yaml:"clickhouse"`
}

func ReadConfigYML(configYML string) *Config {
	once.Do(func() {
		file, err := os.Open(configYML)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
			log.Fatal(err)
		}
		cfg.Database.Password = os.Getenv("DB_PASSWORD")
	})

	printConfig(cfg)

	return cfg
}

func printConfig(cfg *Config) {
	jscfg, _ := json.MarshalIndent(cfg, "", "	")
	logrus.Info(string(jscfg))
	// fmt.Println(string(jscfg))
}
