package main

import (
	"github.com/wnate/Go-000/tree/main/Week03/app"
	"github.com/wnate/Go-000/tree/main/Week03/graceful"
	"log"
	"time"
)

func main() {
	graceful.AddShutDownHook(app.Start(), func() {
		select {
		case <-app.Stop():
			log.Printf("app stop success")
		case <-time.After(30 * time.Second):
			log.Printf("app stop timeout")
		}
	})
}
