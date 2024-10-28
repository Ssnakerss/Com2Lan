package rabbit

import (
	"fmt"
	"rabbit_test/internal/tools"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitChannel struct {
	conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewRabbitChannel(url string) *RabbitChannel {
	var err error
	var rb RabbitChannel
	rb.conn, err = amqp.Dial(url)
	tools.Fail(err, "Failed to connect to RabbitMQ: "+url)
	rb.Ch, err = rb.conn.Channel()
	tools.Fail(err, "Failed to open a channel")
	return &rb
}

func (rb *RabbitChannel) Close() error {
	err := ""
	chErr := rb.Ch.Close()
	if chErr != nil {
		err = chErr.Error()
	}
	connErr := rb.conn.Close()
	if connErr != nil {
		err += "|" + connErr.Error()
	}
	return fmt.Errorf(err)
}
