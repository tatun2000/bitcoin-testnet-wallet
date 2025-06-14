package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	injector := NewKernel(ctx)

	go injector.InjectListenerHandler().Handle()

	<-ctx.Done()
	// graceful shutdown
}
