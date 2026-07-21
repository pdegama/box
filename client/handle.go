package clientapp

import (
	"time"

	"github.com/rellitelink/box/config"
	smtpClient "github.com/rellitelink/box/pkg/client"
	"github.com/rellitelink/box/pkg/logx"
	"gopkg.in/yaml.v3"
)

type EmailHandler struct {
	em         *EmailYAML
	emailYaml  []byte
	amqpStatus *AmqpStatusPublish

	wid    int
	logger *logx.Log
}

type EmailYAML struct {
	Uid       string `yaml:"uid"`
	From      string `yaml:"from"`
	Recipient string `yaml:"recipient"`
	Data      string `yaml:"data"`
}

type EmailStatus struct {
	Time           time.Time
	Uid            string
	Success        bool
	Status         string
	Errors         []smtpClient.ClientServerError
	TempError      bool
	AnyClientError bool
}

func (eh *EmailHandler) handleClient() {
	em, err := openMail(eh.emailYaml)
	if err != nil {
		eh.logger.Error("error on mail parser, worker-%d: %s", eh.wid, err.Error())
		return
	}

	eh.em = em
	eh.logger.Info("Worker %d received a message: %s", eh.wid, eh.em.Uid)

	res := eh.sendMail()

	newRes := EmailStatus{
		Uid:            eh.em.Uid,
		Success:        res.Success,
		Status:         res.Status,
		Time:           res.Time,
		TempError:      res.TempError,
		AnyClientError: res.AnyClientError,
		Errors:         res.Errors,
	}

	errYaml, err := yaml.Marshal(newRes)
	if err != nil {
		eh.logger.Error("error on make status yaml, worker-%d: %s", eh.wid, err.Error())
		return
	}

	eh.amqpStatus.publish(errYaml)
}

func (eh *EmailHandler) sendMail() smtpClient.ClientResponse {
	client := smtpClient.NewClinet()
	client.Logger = eh.logger.GetNewWithPrefix(eh.em.Uid)
	client.Timeout = time.Second * 120
	client.ChunkSize = 1024 * 1000 * 2

	client.SetHostname(config.ConfOpts.Client.HostName)
	client.SetFrom(eh.em.From)
	client.SetRcpt(eh.em.Recipient)

	client.SetData([]byte(eh.em.Data))

	client.StartTls = config.ConfOpts.Tls.StartTls
	client.TlsCert = config.ConfOpts.Tls.Cert
	client.TlsKey = config.ConfOpts.Tls.Key

	client.SendMail()
	res := client.GetResponse()
	return res
}
