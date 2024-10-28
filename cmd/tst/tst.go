package main

import (
	"context"
	"log/slog"
	"os"
	"rabbit_test/internal/comport"
	"time"
)

func main() {
	// create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	port := comport.NewPort("COM1", 9600, logger)
	// defer port.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go port.ListenToWrite(ctx)

	port.WriteData <- []byte("hello") // send data to the port
	time.Sleep(time.Second)
}
