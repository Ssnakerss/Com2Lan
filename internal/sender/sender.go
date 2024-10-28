package sender

import (
	"context"
	"log/slog"
	"rabbit_test/internal/tools"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Sender struct {
	ch       *amqp.Channel
	Name     string
	ChanName string
	logger   *slog.Logger
}

func NewSender(name, chanName string, ch *amqp.Channel, logger *slog.Logger) *Sender {
	s := Sender{
		Name:     name,
		ChanName: chanName,
		ch:       ch,
		logger:   logger,
	}
	err := s.ch.ExchangeDeclare(
		chanName, //name
		"fanout", //type
		true,     //durable
		false,    //autoDelete
		false,    //internal
		false,    //noWait
		nil,
	)
	tools.Fail(err, "exchange declare")
	return &s
}

func (s *Sender) Listen(ctx context.Context, data chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-data:
			s.logger.Debug("sending ", "message", string(data))
			err := s.Send(string(data))
			if err != nil {
				s.logger.Error("send error", "error", err)
			}
		}
	}
}

func (s *Sender) Send(msg string) error {
	return s.ch.Publish(
		s.ChanName, //exchange
		"",         //key
		false,      //mandatory
		false,      //immediate
		amqp.Publishing{
			ReplyTo:     s.Name,
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)
}

func (s *Sender) Delete() error {
	return s.ch.ExchangeDelete(s.ChanName, false, false)
}
