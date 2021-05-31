package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/parithon/minecraftd/minecraft"
	"github.com/parithon/minecraftd/webhooks"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGQUIT)

	go func() {
		s := <-sigs
		log.Printf("RECEIVED: %s", s)
		if err := minecraft.Shutdown(s); err != nil {
			panic(err)
		}
	}()

	minecraft.Startup()
	webhooks.Start()
	minecraft.Wait()
}
