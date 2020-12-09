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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, _ := errgroup.WithContext(ctx)

	// 启动api服务
	group.Go(func() error {
		exitCh, err := serveApi(ctx)
		if err != nil {
			// 通知下所有人都退出
			cancel()
		}
		if exitCh != nil {
			<-exitCh
		}
		return err
	})

	// 启动debug服务
	group.Go(func() error {
		exitCh, err := serveDebug(ctx)
		if err != nil {
			// 通知下所有人都退出
			cancel()
		}
		if exitCh != nil {
			<-exitCh
		}
		return err
	})

	// 启动信号监听服务
	group.Go(func() error {
		err := serveSignal(ctx)
		if err != nil {
			cancel()
		}
		return err
	})

	err := group.Wait()
	if err != nil && !errors.Is(err, errExitByStopSign) {
		log.Printf("\n app shutdown with err, \n %+v", err)
	} else {
		log.Printf("\n app shutdown success")
	}
}

// 如果返回的 exitCh 不为空，需要等待收到通知，才表示服务器完成shutdown
func serveApi(ctx context.Context) (exitCh chan struct{}, err error) {
	addr := ":8080"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, pkgErr.Wrapf(err, "start [api server] fail")
	}
	log.Printf("[api server] luanch success, running on %s", l.Addr())

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// 自己curl请求，然后再ctrl+c关闭服务器
		// 模拟关闭服务器时，还有正在处理中的请求
		time.Sleep(10 * time.Second)
		fmt.Fprintln(resp, "Hello world!")
	})

	exitCh = make(chan struct{})

	go func() {
		defer func() {
			// 当调了shutdown()方法后，外面的server.Serve(l) 会异步立即返回，
			// 所以这里需要等shutdown处理完后，再通知下外面的调用者
			close(exitCh)
		}()
		select {
		case <-ctx.Done():
			timeoutCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
			server.Shutdown(timeoutCtx)
		}
	}()

	return exitCh, server.Serve(l)
}

// 如果返回的 exitCh 不为空，需要等待收到通知，才表示服务器完成shutdown
func serveDebug(ctx context.Context) (exitCh chan struct{}, err error) {
	addr := "127.0.0.1:8081"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, pkgErr.Wrapf(err, "start [debug server] fail")
	}
	log.Printf("[dubug server] luanch success, running on %s", l.Addr())
	server := &http.Server{Handler: http.DefaultServeMux}

	exitCh = make(chan struct{})
	go func() {
		defer func() {
			close(exitCh)
		}()
		select {
		case <-ctx.Done():
			timeoutCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
			server.Shutdown(timeoutCtx)
		}
	}()
	return exitCh, server.Serve(l)
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
