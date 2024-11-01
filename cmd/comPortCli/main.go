package main

import (
	"com2rabbitmq/internal/app"
	"time"

	"context"
	"log"
	"log/slog"
)

func main() {

	var acfg app.Config
	acfg.Read()

	if acfg.PortName == "" || acfg.ClientId == "" || acfg.ChannelName == "" {
		log.Fatal("port-name, client-id, channel-name are required")
	}
	// create a logger
	logger := app.Logger(slog.LevelDebug)

	mctx, mcancel := context.WithCancel(context.Background()) // create a context for the main goroutine
	defer mcancel()

	go app.CtrlC(mctx, mcancel, logger)
	for {
		logger.Debug("start working loop")

		wrtsig := make(chan struct{})
		rdrsig := make(chan struct{})

		wctx, wcancel := context.WithCancel(mctx) //create a context for work loop
		defer wcancel()

		rb, wrt, rdr, port := app.Connect(wctx, &acfg, logger)
		if rb == nil {
			logger.Error("connection procedure interrupted, app will quit") // can't connect to port
			return
		}
		defer rb.Close()

		go port.ComToQueue(wctx)
		go port.QueueToCom(wctx)
		go wrt.Listen(wctx, port.ReadData, wrtsig)
		go rdr.Listen(wctx, port.WriteData, rdrsig)

		select {
		case <-mctx.Done():
			wcancel()
			time.Sleep(time.Second)
			logger.Info("all done, bye-bye")
			return
		case <-wrtsig:
			wcancel()
			time.Sleep(time.Second)
			//reconnecting
		case <-rdrsig:
			wcancel()
			time.Sleep(time.Second)
			//reconnecting
		}
		logger.Warn("reconnecting")
	}
}
