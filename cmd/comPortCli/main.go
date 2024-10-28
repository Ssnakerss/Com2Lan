package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"rabbit_test/internal/comport"
	"rabbit_test/internal/rabbit"
	"rabbit_test/internal/reader"
	"rabbit_test/internal/sender"
	"syscall"
)

type AppConfig struct {
	PortName    string
	ClientId    string
	ChannelName string
	RabbitURL   string
}

func (acfg *AppConfig) Read() {
	flag.StringVar(&acfg.PortName, "port-name", "", "port name")
	flag.StringVar(&acfg.ClientId, "client-id", "", "client id")
	flag.StringVar(&acfg.ChannelName, "channel-name", "", "channel name")
	flag.StringVar(&acfg.RabbitURL, "rabbit-url", "amqp://guest:guest@localhost:5672/", "rabbit url")
	flag.Parse()

}

func fail(err error, message string) {
	if err != nil {
		log.Fatal(message, "err:", err)
	}
}

func main() {
	// create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	var acfg AppConfig
	acfg.Read()
	if acfg.PortName == "" || acfg.ClientId == "" || acfg.ChannelName == "" {
		log.Fatal("port-name, client-id, channel-name are required")
	}
	mctx, mcancel := context.WithCancel(context.Background()) // create a context for the main goroutine
	defer mcancel()

	rb := rabbit.NewRabbitChannel(acfg.RabbitURL)
	defer rb.Close()

	snd := sender.NewSender(acfg.ClientId, acfg.ChannelName, rb.Ch, logger)
	rdr := reader.NewReader(acfg.ClientId, acfg.ChannelName, rb.Ch, logger)
	port := comport.NewPort(acfg.PortName, 9600, logger)
	// defer port.Close() // close the port

	//-----------------------------------------------------------------
	go func() {
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
			mcancel()
		case <-mctx.Done():
			logger.Info("shutting down")
		}
	}()

	go port.ListenToWrite(mctx)
	go port.ListenToRead(mctx)

	go snd.Listen(mctx, port.ReadData)
	go rdr.Listen(mctx, port.WriteData)

	// for {
	// 	select {
	// 	case <-mctx.Done():
	// 		return
	// 	case message := <-data:
	// 		logger.Info("got new message", "message", message) // receive a message
	// 		port.Write(message)
	// 	}

	// }
	<-mctx.Done()
}
