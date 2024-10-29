package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"rabbit_test/internal/comport"
	"rabbit_test/internal/rabbit"
	"rabbit_test/internal/reader"
	"rabbit_test/internal/sender"
	"rabbit_test/internal/tools"
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
	go tools.CtrlC(mctx, mcancel, logger)

	go port.ListenToWrite(mctx)
	go port.ListenToRead(mctx)

	go snd.Listen(mctx, port.ReadData)
	go rdr.Listen(mctx, port.WriteData)

	<-mctx.Done()
	logger.Info("all done, bye-bye")
}
