package config

import (
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var gConfig = &atomic.Value{}

type Config struct {
	MYSQL struct {
		Host            string        `yaml:"host"`
		Port            string        `yaml:"port"`
		UserName        string        `yaml:"username"`
		PassWord        string        `yaml:"password"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		MaxOpenConns    int           `yaml:"max_open_conns"`
	}
	REDIS struct {
		Host           string `yaml:"host"`
		Port           string `yaml:"port"`
		PassWord       string `yaml:"password"`
		PoolSize       int    `yaml:"pool_size"`
		TokenMaxExTime int    `yaml:"token_max_expired_time"`
	}
	HTTP struct {
		Port           string `yaml:"port"`
		StaticFilePath string `yaml:"static_file_path"`
	}
	RPC struct {
		Address      string `yaml:"address"`
		ConnPoolSize int    `yaml:"conn_pool_size"`
	}
	DefaultImagePath string `yaml:"default_image_path"`
}

func GetConfig() *Config {
	if configVar, ok := gConfig.Load().(*Config); ok {
		return configVar
	}
	return nil
}

// InitConfig 从配置平台拉取配置并定时更新
func InitConfig() {
	LoadAndWatch(ReadConfig)
}

// 解析配置内容
func ReadConfig(content string) error {
	yamlFile, err := ioutil.ReadFile(content)
	if err != nil {
		fmt.Errorf("[ReadConfig]read yaml failed:%+v", err)
		return err
	}

	//fmt.Printf("[ReadConfig]read success:%+v", string(yamlFile))
	config := &Config{}

	unmarshalErr := yaml.Unmarshal(yamlFile, config)
	if unmarshalErr != nil {
		fmt.Errorf("[ReadConfig]unmarshal failed:%+v", unmarshalErr)
		return unmarshalErr
	}

	//fmt.Printf("[ReadConfig]unmarshal success:%+v", config)

	gConfig.Store(config)
	return nil
}

// 主动刷新+周期刷新
func LoadAndWatch(handle func(string) error) {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Errorf("[LoadAndWatch]read yaml failed:%+v", err)
		panic(err)
	}

	handle(string(yamlFile)) // 刷新配置

	go func() {
		ticket := time.NewTicker(10 * time.Second)
		for _ = range ticket.C {
			yamlFile, err := ioutil.ReadFile("config.yaml")
			if err != nil {
				fmt.Errorf("[LoadAndWatch]read yaml failed:%+v", err)
				continue
			}

			handle(string(yamlFile)) // 刷新配置
		}
	}()

	return
}
