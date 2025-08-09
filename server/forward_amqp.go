package serverapp

import (
	"fmt"
	"log"
	"net/url"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rellitelink/box/config"
	"github.com/rellitelink/box/pkg/logx"
	"github.com/rellitelink/box/pkg/server"
)

type MailFwdBackendAmqp struct {
	server.ForwardBackend

	uri     string
	queue   amqp.Queue
	channel *amqp.Channel
	logger  *logx.Log
}

func (mailFwd *MailFwdBackendAmqp) Init() {

	// load config
	encodedPassword := url.QueryEscape(config.ConfOpts.Amqp.Password)
	mailFwd.uri = fmt.Sprintf("amqp://%s:%s@%s:%d/", config.ConfOpts.Amqp.Username, encodedPassword, config.ConfOpts.Amqp.Host, config.ConfOpts.Amqp.Port)
	client, err := amqp.Dial(mailFwd.uri)
	if err != nil {
		mailFwd.logger.Error("AMQP Connection Faild...")
		if config.ConfOpts.Dev {
			fmt.Println(err)
		}
		os.Exit(1)
	} else {
		mailFwd.logger.Info("AMQP Connection Success")
	}

	channel, err := client.Channel()
	if err != nil {
		mailFwd.logger.Error("Failed to open a AMQP channel...")
		if config.ConfOpts.Dev {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	mailFwd.channel = channel

	queueName := config.ConfOpts.Amqp.Queue
	if queueName == "" {
		queueName = config.ConfOpts.HostName
	}

	queue, err := channel.QueueDeclare(
		queueName, // queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		mailFwd.logger.Error("Failed to declare a AMQP queue...")
		if config.ConfOpts.Dev {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	mailFwd.queue = queue

	log.Println("Init AMQP Forward method")
}

func (mailFwd *MailFwdBackendAmqp) ForwardMail(email server.Email) {
	data, err := email.ToDocument()
	if err != nil {
		log.Println("error to gen document")
	}

	fmt.Println(string(data))

	err = mailFwd.channel.Publish(
		"",                 // exchange
		mailFwd.queue.Name, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)

	if err != nil {
		mailFwd.logger.Error("Failed to publish a message")
		if config.ConfOpts.Dev {
			fmt.Println(err)
		}
		return
	}

	fmt.Println("email add successful")

	emailS, ok := email.ParseMail()
	if !ok {
		mailFwd.logger.Error("Error In mail Parse")
		return
	}

	fmt.Println("email: ")
	fmt.Println(emailS)
}

func (mailFwd *MailFwdBackendAmqp) ExistMailBox(rcpt string) bool {
	return checkMaiboxFromRcpt(rcpt)
}
