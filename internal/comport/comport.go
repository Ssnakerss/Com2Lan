package comport

import (
	"bufio"
	"context"
	"log/slog"

	"github.com/tarm/serial"
)

type Com struct {
	name      string
	baud      int
	ReadData  chan []byte
	WriteData chan []byte
	Port      *serial.Port
	logger    *slog.Logger
}

func New(comPortName string, baud int, logger *slog.Logger) (*Com, error) {
	c := &serial.Config{Name: comPortName, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}

	com := Com{
		name:      comPortName,
		baud:      baud,
		ReadData:  make(chan []byte),
		WriteData: make(chan []byte),
		Port:      s,
		logger:    logger,
	}
	return &com, nil
}

func (c *Com) ComToQueue(ctx context.Context) {
	c.logger.Info("start listen to read", "port", c.name)
	reader := bufio.NewReader(c.Port)
	for {
		select {
		case <-ctx.Done():
			c.logger.Warn("ComToQueue stop listen ", "port", c.name)
			return
		default:
			data, err := reader.ReadBytes(0x03)
			if err != nil {
				c.logger.Warn("read error", slog.String("err", err.Error()))
			} else {
				if data != nil {
					c.logger.Info("read", "port", c.name, "data", string(data))
					c.ReadData <- data //TODO
				}
			}
		}
	}
}

func (c *Com) QueueToCom(ctx context.Context) {
	c.logger.Info("start listen to write", "port", c.name)
	for {
		select {
		case <-ctx.Done():
			c.logger.Warn("QueueToCom stop listen ", "port", c.name)
			return
		case data := <-c.WriteData:
			c.logger.Info("write", "port", c.name, "data", string(data))
			err := c.write(data)
			if err != nil {
				c.logger.Warn("write error", slog.String("err", err.Error()))
			}
		}
	}
}

func (c *Com) write(data []byte) error {
	_, err := c.Port.Write(data)
	return err
}

func (c *Com) Close() error {
	return c.Port.Close()
}
