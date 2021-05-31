package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/parithon/minecraftd/minecraft"
	"github.com/parithon/minecraftd/webhooks"
)

func create_lock_file(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
}

func main() {

	lockfile := fmt.Sprintf("%s.lock", os.Args[0])

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "shutdown":
			{
				webhooks.Shutdown(false)
				return
			}
		case "terminate":
			{
				webhooks.Shutdown(true)
				return
			}
		}
	}

	if _, err := create_lock_file(lockfile); err != nil {
		fmt.Println(err)
	}

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
