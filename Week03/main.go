package main

import (
	"context"
	"errors"
	"fmt"
	pkgErr "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var errExitByStopSign = errors.New("err: receive stop sign")

func main() {
	ctx, _ := context.WithCancel(context.Background())
	group, _ := errgroup.WithContext(context.Background())

	apiServer := newApiServer()
	debugServer := newDebugServer()

	// 启动api服务
	group.Go(func() error {
		addr := ":8080"
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return pkgErr.Wrapf(err, "start [api server] fail")
		}
		log.Printf("[api server] luanch success, running on %s", l.Addr())
		return apiServer.Serve(l)
	})

	// 启动debug服务
	group.Go(func() error {
		addr := "127.0.0.1:8081"
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return pkgErr.Wrapf(err, "start [debug server] fail")
		}
		log.Printf("[dubug server] luanch success, running on %s", l.Addr())
		return debugServer.Serve(l)
	})

	// 启动信号监听服务
	group.Go(func() error {
		err := serveSignal(ctx)
		if err != nil && errors.Is(err, errExitByStopSign) {
			apiServer.Shutdown(context.Background())
			debugServer.Shutdown(context.Background())
		}
		return err
	})

	err := group.Wait()
	if err != nil {
		log.Printf("\n app shutdown with err, \n %+v", err)
	} else {
		log.Printf("\n app shutdown success")
	}

}

// 新建一个api服务器对象
func newApiServer() *http.Server {
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// 模拟有处理中的请求
		time.Sleep(10 * time.Second)
		fmt.Fprintln(resp, "Hello world!")
	})
	return server
}

// 新建一个debug server
func newDebugServer() *http.Server {
	server := &http.Server{Handler: http.DefaultServeMux}
	return server
}

func serveSignal(ctx context.Context) error {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	log.Printf("[shutdown hook] bind success")
	select {
	case <-signalChan:
		return errExitByStopSign
	case <-ctx.Done():
		return nil
	}
}
