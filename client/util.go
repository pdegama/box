package clientapp

import (
	"fmt"
	"os"

	"github.com/pdegama/box/config"
	"github.com/pdegama/box/pkg/logx"
	"gopkg.in/yaml.v3"
)

func makeLogger() *logx.Log {
	logFile := os.Stdout
	if !config.ConfOpts.Dev {
		var err error
		logFile, err = os.OpenFile(config.ConfOpts.Client.LogDirPath+"/"+config.ConfOpts.Client.LogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logx.LogError("Error opening file:", err)
			return nil
		}
	}
	logger := logx.NewLoggerWithPrefix(logFile, "CLIENT")
	return logger
}

func openMail(data []byte) (*EmailYAML, error) {
	e := new(EmailYAML)
	err := yaml.Unmarshal(data, e)
	if err != nil {
		return nil, err
	}

	if e.Uid == "" {
		return nil, fmt.Errorf("email uid not found")
	}

	if e.From == "" {
		return nil, fmt.Errorf("email from field not found")
	}

	if e.Recipient == "" {
		return nil, fmt.Errorf("email recipient field not found")
	}

	if e.Data == "" {
		return nil, fmt.Errorf("email data not found")
	}

	return e, nil
}
