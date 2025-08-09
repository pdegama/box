package config

import (
	"os"

	"github.com/pdegama/box/pkg/logx"
	"gopkg.in/yaml.v3"
)

// loaded config
var ConfOpts *ConfigsOptions = nil

func LoadConfig() {
	if ConfOpts == nil {
		ConfOpts = GetConfig()
	}

	postConfigAction()
}

func GetConfig() *ConfigsOptions {
	config := defaultConfig()

	notAnyConfigFile, configData := getConfigFile()
	if !notAnyConfigFile {
		err := yaml.Unmarshal(configData, &config)
		if err != nil {
			logx.LogError("failed to parse config file", err)
			os.Exit(1)
		}
	}

	return &config
}

var configFiles []string = []string{
	"/etc/box.yaml",
	"/etc/box.yml",
	"./config.yaml",
	"./config.yml",
}

// return notFound and file data if
// not notFount is true than file data
// is empty string
func getConfigFile() (bool, []byte) {
	for _, configFile := range configFiles {
		fileBytes, err := os.ReadFile(configFile)
		if err == nil {
			return false, fileBytes
		}
	}
	return true, []byte{}
}
