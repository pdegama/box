package config

import (
	"os"

	"github.com/pdegama/box/pkg/logx"
)

func postConfigAction() {

	// make log dir if not exists
	_, err := os.Stat(ConfOpts.LogDirPath)
	// fmt.Println(err)
	if err != nil {

		if os.IsNotExist(err) {
			err := os.MkdirAll(ConfOpts.LogDirPath, 0755)
			if err != nil {
				logx.LogError(err.Error(), err)
				os.Exit(1)
			}
		} else {
			logx.LogError(err.Error(), err)
			os.Exit(1)
		}
	}

	// if queue name is empty
	if ConfOpts.Amqp.Queue == "" {
		ConfOpts.Amqp.Queue = ConfOpts.HostName
	}
	if ConfOpts.Amqp.StatusQueue == "" {
		ConfOpts.Amqp.StatusQueue = ConfOpts.HostName + "-status"
	}

	// set default config for clinet
	if ConfOpts.Client.HostName == "" {
		ConfOpts.Client.HostName = ConfOpts.HostName
	}

	if ConfOpts.Client.LogDirPath == "" {
		ConfOpts.Client.LogDirPath = ConfOpts.LogDirPath
	}

	if ConfOpts.Client.LogFilePath == "" {
		ConfOpts.Client.LogFilePath = ConfOpts.LogFilePath
	}

	if ConfOpts.Client.Amqp.Host == "" {
		ConfOpts.Client.Amqp.Host = ConfOpts.Amqp.Host
	}

	if ConfOpts.Client.Amqp.Port == 0 {
		ConfOpts.Client.Amqp.Port = ConfOpts.Amqp.Port
	}

	if ConfOpts.Client.Amqp.Username == "" {
		ConfOpts.Client.Amqp.Username = ConfOpts.Amqp.Username
	}

	if ConfOpts.Client.Amqp.Password == "" {
		ConfOpts.Client.Amqp.Password = ConfOpts.Amqp.Password
	}

	if ConfOpts.Client.Amqp.Queue == "" {
		ConfOpts.Client.Amqp.Queue = ConfOpts.Amqp.Queue
	}

	if ConfOpts.Client.Amqp.StatusQueue == "" {
		ConfOpts.Client.Amqp.StatusQueue = ConfOpts.Amqp.StatusQueue
	}

}
