package serverapp

import (
	"os"
	"sync"
	"time"

	"github.com/rellitelink/box/config"
	"github.com/rellitelink/box/pkg/logx"
	"github.com/rellitelink/box/pkg/server"
)

func StartServer() {

	server.SetConfig(server.SMTPConfig{
		Name:     config.ConfOpts.Name,
		HostName: config.ConfOpts.HostName,
		ESMTP: server.ESMTPOptions{
			Enable:      true,
			Tls:         config.ConfOpts.Tls.StartTls,
			Utf8:        true,
			BinaryMime:  true,
			MessageSize: config.ConfOpts.MessageSize,
		},

		ClientGreet: "Hello, Client!",
		ClientByyy:  "Ok, Byyy!",

		SpfCheck:      config.ConfOpts.SpfCheck,
		MaxRecipients: config.ConfOpts.MaxRecipients,
		MaxClients:    config.ConfOpts.MaxClients,

		CheckMailBoxExist: config.ConfOpts.CheckMailBoxExist,
		TlsConfig: server.TlsKeyCert{
			Key:  config.ConfOpts.Tls.Key,
			Cert: config.ConfOpts.Tls.Cert,
		},

		Timeout: time.Minute * 5,
		Dev:     config.ConfOpts.Dev,
	})

	// logger
	logFile := os.Stdout
	if !config.ConfOpts.Dev {
		var err error
		logFile, err = os.OpenFile(config.ConfOpts.LogDirPath+"/"+config.ConfOpts.LogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logx.LogError("Error opening file:", err)
			return
		}
	}
	logger := logx.NewLogger(logFile)

	server.SetMailFwdMethod(&MailFwdBackendAmqp{
		logger: logger,
	})

	serverWait := sync.WaitGroup{}

	// IPv4 listen
	serverWait.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		smtpServer := server.SMTPServer{
			Host:            "0.0.0.0",
			Port:            config.ConfOpts.Port,
			ServerWaitGroup: wg,
		}

		// fmt.Println(smtpServer)

		smtpServer.SetLogger(logger)
		smtpServer.Listen()
		smtpServer.AcceptConnections()

	}(&serverWait)

	// IPv6 listen
	serverWait.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		smtpServer := server.SMTPServer{
			Host:            "::",
			Port:            config.ConfOpts.Port,
			ServerWaitGroup: wg,
			IsIPv6:          true,
		}

		smtpServer.SetLogger(logger)
		smtpServer.Listen()
		smtpServer.AcceptConnections()

	}(&serverWait)

	serverWait.Wait()
}
