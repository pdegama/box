package clientapp

import (
	"fmt"
	"net/url"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rellitelink/box/config"
	"github.com/rellitelink/box/pkg/logx"
)

func getAmqpConn(logger *logx.Log) *amqp.Connection {
	encodedPassword := url.QueryEscape(config.ConfOpts.Client.Amqp.Password)
	amqpUri := fmt.Sprintf("amqp://%s:%s@%s:%d/", config.ConfOpts.Client.Amqp.Username, encodedPassword, config.ConfOpts.Client.Amqp.Host, config.ConfOpts.Client.Amqp.Port)
	conn, err := amqp.Dial(amqpUri)
	if err != nil {
		logger.Error("error on amqp conn: %s", err)
		os.Exit(1)
	} else {
		logger.Info("AMQP Connection Success")
	}

	return conn
}

func getAmqpConsume(logger *logx.Log, conn *amqp.Connection, queueName string) <-chan amqp.Delivery {
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("error on amqp channel: %s", err)
		os.Exit(1)
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		logger.Error("error on amqp queue declare: %s", err)
		os.Exit(1)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logger.Error("error on amqp consume: %s", err)
		os.Exit(1)
	}

	return msgs
}

type AmqpStatusPublish struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	q         amqp.Queue
	queueName string
	logger    *logx.Log
}

func (a *AmqpStatusPublish) setChannel() {
	ch, err := a.conn.Channel()
	if err != nil {
		a.logger.Error("error on amqp channel in status: %s", err)
		os.Exit(1)
	}
	// Declare a queue (you can also use an existing queue)
	q, err := ch.QueueDeclare(
		a.queueName, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		a.logger.Error("error on amqp queue declare in status: %s", err)
		os.Exit(1)
	}

	a.q = q
	a.ch = ch
}

func (a *AmqpStatusPublish) publish(msg []byte) {
	err := a.ch.Publish(
		"",       // exchange
		a.q.Name, // routing key (queue name)
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		a.logger.Error("error on amqp publish: %s", err)
	}
}
