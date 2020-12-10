package main

import (
	"github.com/wnate/Go-000/tree/main/Week03/app"
	"github.com/wnate/Go-000/tree/main/Week03/graceful"
)

func main() {
	graceful.AddShutDownHook(app.Start(), func() {
		<-app.Stop()
	})
}
