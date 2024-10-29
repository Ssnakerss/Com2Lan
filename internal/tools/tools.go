package tools

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Fail(err error, message string) {
	if err != nil {
		log.Fatal(message, "err:", err)
	}
}

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
