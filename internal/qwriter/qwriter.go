package qwriter

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QWriter struct {
	ch       *amqp.Channel
	Name     string
	ChanName string
	logger   *slog.Logger
}

func New(name, chanName string, ch *amqp.Channel, logger *slog.Logger) (*QWriter, error) {
	s := QWriter{
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
	return &s, err
}

func (s *QWriter) Listen(ctx context.Context, data chan []byte, sig chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Warn("context canceled, queuewriter stop listen ", "channel", s.Name)
			return
		case data, ok := <-data:
			if !ok {
				s.logger.Warn("input channel closed, queuewriter stop listen ", "channel", s.Name)
				close(sig)
				return
			}

			s.logger.Debug("sending ", "message", string(data))
			err := s.Send(string(data))
			if err != nil {
				s.logger.Error("queuewriter send error", "channel", s.Name, "error", err)
				close(sig)
				return
			}
		}
	}
}

func (s *QWriter) Send(msg string) error {
	return s.ch.Publish(
		s.ChanName, //exchange
		"",         //key
		false,      //mandatory
		false,      //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Headers: amqp.Table{
				"ReplyTo": s.Name,
			},
			Body: []byte(msg),
		},
	)
}

func (s *QWriter) Delete() error {
	return s.ch.ExchangeDelete(s.ChanName, false, false)
}
