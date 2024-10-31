package rabbit

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitChannel struct {
	conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewRabbitChannel(url string) (*RabbitChannel, error) {
	var err error
	var rb RabbitChannel

	rb.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	rb.Ch, err = rb.conn.Channel()
	return &rb, err
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
