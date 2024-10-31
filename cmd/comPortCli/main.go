package main

import (
	"com2rabbitmq/internal/app"
	"time"

	"context"
	"log"
	"log/slog"

	"golang.org/x/sync/errgroup"
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

	// wctx, wcancel := context.WithCancel(mctx) //create a context for work loop
	// defer wcancel()

	sig := make(chan struct{})

	for {
		rb, wrt, rdr, port := app.Connect(&acfg, logger)
		defer rb.Close()

		g, ctx := errgroup.WithContext(mctx)

		g.Go(port.ComToQueue(ctx))
		g.Go(port.QueueToCom(ctx))
		g.Go(wrt.Listen(ctx, port.ReadData))
		g.Go(rdr.Listen(ctx, port.WriteData))

		go func() {
			err := g.Wait()
			if err != nil {
				sig <- struct{}{}
			}
		}()

		select {
		case <-mctx.Done():
			time.Sleep(time.Second)
			logger.Info("all done, bye-bye")
			return
		case <-sig:
			//reconnecting
		}
	}

}
