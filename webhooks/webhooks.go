package webhooks

import (
	"fmt"
	"log"
	"net/http"
	"syscall"
)

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Shutdown request received. Shutting down in 30 seconds...\n")
	syscall.Kill(syscall.Getpid(), syscall.SIGQUIT)
}

func shutdownNowHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Shutdown request received. Shutting down server NOW...\n")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}

func msgHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Sending message to players\n")
	// TODO: Add logic to send message to players
}

func Start() {

	http.HandleFunc("/webhooks/shutdown", shutdownHandler)
	http.HandleFunc("/webhooks/shutdown/now", shutdownNowHandler)
	http.HandleFunc("/webhooks/msg", msgHandler)

	log.Println("Starting webhooks...")
	go func() {
		log.Fatal(http.ListenAndServe(":8090", nil))
	}()
	log.Println("Webhooks started")

}

func Shutdown(now bool) error {
	if !now {
		log.Println("Shutting down in 30 seconds...")
		if _, err := http.Get("http://localhost:8090/webhooks/shutdown"); err != nil {
			return err
		}
	} else {
		log.Println("Shutting down NOW...")
		if _, err := http.Get("http://localhost:8090/webhooks/shutdown/now"); err != nil {
			return err
		}
	}

	return nil
}
