package reader

import (
	"context"
	"log/slog"
	"rabbit_test/internal/tools"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Reader struct {
	ch       *amqp.Channel
	Name     string
	ChanName string
	queue    amqp.Queue
	Messages <-chan amqp.Delivery //возвращает все сообщения которые пришли в очередь
	logger   *slog.Logger
}

func NewReader(name, chanName string, ch *amqp.Channel, logger *slog.Logger) *Reader {
	r := Reader{
		ch:       ch,
		Name:     name,
		ChanName: chanName,
		logger:   logger,
	}
	var err error
	r.queue, err = ch.QueueDeclare(
		"",    //name если не указано - генерится автоматом
		false, //durable
		false, //delete when unused
		true,  //exclusive
		false, //no-wait
		nil,   //arguments
	)
	tools.Fail(err, name+":queue")

	err = ch.QueueBind(
		r.queue.Name,
		"",
		chanName,
		false,
		nil,
	)
	tools.Fail(err, name+": queue bind")

	r.Messages, err = ch.Consume(
		r.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	tools.Fail(err, name+": get messages")

	return &r
}

func (r *Reader) Listen(ctx context.Context, data chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-r.Messages:
			if message.ReplyTo != r.Name {
				r.logger.Debug("received message", "message", message.Body, "from", message.ReplyTo)
				data <- []byte(message.Body)
			}
		}
	}
}
