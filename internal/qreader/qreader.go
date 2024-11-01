package qreader

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QReader struct {
	ch       *amqp.Channel
	Name     string
	ChanName string
	queue    amqp.Queue
	Messages <-chan amqp.Delivery //возвращает все сообщения которые пришли в очередь
	logger   *slog.Logger
}

func New(name, chanName string, ch *amqp.Channel, logger *slog.Logger) (*QReader, error) {
	r := QReader{
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
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		r.queue.Name,
		"",
		chanName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	r.Messages, err = ch.Consume(
		r.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	return &r, err
}

func (r *QReader) Listen(ctx context.Context, data chan []byte, sig chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			r.logger.Warn("context canceled, queuereader stop listen ", "port", r.Name)
			return
		case message, ok := <-r.Messages:
			if !ok {
				r.logger.Warn("queue message channel closed, qreader stop listen ", "channel", r.Name)
				close(sig) //оповещаем все что произошла ошибка чтения из очереди
				return
			}
			if message.Headers["ReplyTo"] != r.Name {
				r.logger.Debug("received message", "message", message.Body, "from", message.Headers["ReplyTo"])
				data <- []byte(message.Body)
			}
		}
	}
}
