package app

import (
	"com2rabbitmq/internal/comport"
	"com2rabbitmq/internal/qreader"
	"com2rabbitmq/internal/qwriter"
	"com2rabbitmq/internal/rabbit"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CtrlC(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	select {
	case s := <-exit:
		logger.Info("received signal: ", "syscal", s.Signal)
		logger.Info("shutting down")
		cancel()
	case <-ctx.Done():
		logger.Info("shutting down")
	}
}

func Connect(acfg *Config, logger *slog.Logger) (
	*rabbit.RabbitChannel,
	*qwriter.QWriter,
	*qreader.QReader,
	*comport.Com) {

	var rb *rabbit.RabbitChannel
	var wrt *qwriter.QWriter
	var rdr *qreader.QReader
	var port *comport.Com

	err := fmt.Errorf("dummy")
	for err != nil {
		rb, err = rabbit.NewRabbitChannel(acfg.RabbitURL)
		if err != nil {
			logger.Error("rabbit", "error", err.Error())
		} else {
			logger.Info("rabbit connected")

			wrt, err = qwriter.New(acfg.ClientId, acfg.ChannelName, rb.Ch, logger)
			if err != nil {
				logger.Error("qwriter", "error", err.Error())
			} else {
				logger.Info("qwriter created")
			}

			rdr, err = qreader.New(acfg.ClientId, acfg.ChannelName, rb.Ch, logger)
			if err != nil {
				logger.Error("qreader", "error", err.Error())
			} else {
				logger.Info("qreader created")
			}
		}
		if err != nil {
			//попробуем еще раз после паузы
			time.Sleep(time.Second * 5)
		}
	}

	err = fmt.Errorf("dummy")
	for err != nil {
		port, err = comport.New(acfg.PortName, 9600, logger)
		if err != nil {
			logger.Error("com port open", "error", err)
		}
	}

	logger.Info("com port opened")
	return rb, wrt, rdr, port
}
