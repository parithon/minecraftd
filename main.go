package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/parithon/minecraftd/management"
	"github.com/parithon/minecraftd/minecraft"
	"github.com/parithon/minecraftd/utils"
)

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "shutdown":
			{
				management.Shutdown(false)
				return
			}
		case "terminate":
			{
				management.Shutdown(true)
				return
			}
		case "healthcheck":
			{
				management.Healthcheck()
				return
			}
		}
	}

	lockfile := fmt.Sprintf("%s.lock", os.Args[0])
	if _, err := utils.CreateLock(lockfile); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGQUIT)

	go func() {
		s := <-sigs
		if err := minecraft.Shutdown(s); err != nil {
			utils.Fatal(err)
		}
	}()

	management.Start()
	minecraft.Startup()
	minecraft.Wait()
}
