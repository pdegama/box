// U2SMTP - config
//
// Licensed under the MIT License for individual use or a commercial
// license for enterprise use. See LICENSE.txt and COMMERCIAL_LICENSE.txt.

package config

import (
	"os"

	"github.com/p4rthk4/u2smtp/pkg/logx"
	"gopkg.in/yaml.v3"
)

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
	"/etc/u2/smtp/config",
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
