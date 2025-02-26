package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/utmstack/UTMStack/installer/utils"
)

type Config struct {
	MainServer    string `yaml:"main_server"`
	Branch        string `yaml:"branch"`
	Password      string `yaml:"password"`
	DataDir       string `yaml:"data_dir"`
	ServerType    string `yaml:"server_type"`
	ServerName    string `yaml:"server_name"`
	InternalKey   string `yaml:"internal_key"`
	UpdatesFolder string `yaml:"updates_folder"`
}

var (
	config     *Config
	configOnce sync.Once
)

func GetConfig() *Config {
	configOnce.Do(func() {
		config = &Config{}
		if utils.CheckIfPathExist(ConfigPath) {
			err := utils.ReadYAML(ConfigPath, config)
			if err != nil {
				fmt.Printf("error reading config file: %v", err)
				os.Exit(1)
			}
		}

		mainIP, err := utils.GetMainIP()
		if err != nil {
			fmt.Printf("error getting main IP: %v", err)
			os.Exit(1)
		}
		config.MainServer = mainIP

		sName, err := os.Hostname()
		if err != nil {
			fmt.Printf("error getting hostname: %v", err)
			os.Exit(1)
		}
		config.ServerName = sName

		if config.ServerType != "aio" &&
			config.ServerType != "cloud" {
			config.ServerType = "aio"
		}

		if config.Password == "" {
			config.Password = utils.GenerateSecret(16)
		}

		if config.InternalKey == "" {
			config.InternalKey = utils.GenerateSecret(32)
		}

		if config.Branch != "alpha" &&
			config.Branch != "beta" &&
			config.Branch != "rc" {
			config.Branch = "prod"
		}

		if config.DataDir == "" {
			config.DataDir = "/utmstack"
		}

		if config.UpdatesFolder == "" {
			config.UpdatesFolder = utils.MakeDir(0777, config.DataDir, "updates")
		}

		err = config.Set()
		if err != nil {
			fmt.Printf("error setting config: %v", err)
			os.Exit(1)
		}
	})

	return config
}

func (c *Config) Set() error {
	return utils.WriteYAML(ConfigPath, c)
}
