package clientapp

import (
	"github.com/pdegama/box/config"
)

func StartClient() {
	logger := makeLogger()

	amqlConn := getAmqpConn(logger)

	amqlStatus := AmqpStatusPublish{
		conn:      amqlConn,
		queueName: config.ConfOpts.Client.Amqp.StatusQueue,
		logger:    logger,
	}
	amqlStatus.setChannel()

	emails := getAmqpConsume(logger, amqlConn, config.ConfOpts.Client.Amqp.Queue)

	forever := make(chan bool)
	for i := 0; i < config.ConfOpts.Client.Worker; i++ {
		go func(workerID int) {
			for d := range emails {
				emailHandler := EmailHandler{
					emailYaml:  d.Body,
					amqpStatus: &amqlStatus,
					logger:     logger,
					wid:        workerID,
				}
				emailHandler.handleClient()
			}
		}(i)
	}

	<-forever
}
