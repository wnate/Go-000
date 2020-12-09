package app

import (
	"context"
	"fmt"
	pkgErr "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	stopCh   chan struct{}
	exitCh   chan struct{}
	stopOnce sync.Once
)

func Start() context.Context {
	stopCh = make(chan struct{})
	exitCh = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	group, _ := errgroup.WithContext(ctx)

	group.Go(func() error {
		stopCh, err := serveApi(ctx)
		if err != nil {
			cancel()
		}
		if stopCh != nil {
			// 等待通知退出
			<-stopCh
		}
		return err
	})

	group.Go(func() error {
		stopCh, err := serveDebug(ctx)
		if err != nil {
			cancel()
		}
		if stopCh != nil {
			// 等待通知退出
			<-stopCh
		}
		return err
	})

	group.Go(func() error {
		select {
		case <-stopCh:
			cancel()
		case <-ctx.Done():
			break
		}
		return nil
	})

	go func() {
		defer func() {
			// 通知所有子任务都安全退出了，系统可以放心退出
			close(exitCh)
		}()
		if err := group.Wait(); err != nil {
			log.Printf("app start err %+v", err)
		}
	}()

	return ctx
}

func Stop() {
	stopOnce.Do(func() {
		close(stopCh)
	})
	// 等待通知可以退出
	<-exitCh
}

// 如果返回的 exitCh 不为空，需要等待收到通知，才表示服务器完成shutdown
func serveApi(ctx context.Context) (stopCh chan struct{}, err error) {
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

	stopCh = make(chan struct{})

	go func() {
		defer func() {
			// 通知shutdown完成了
			close(stopCh)
		}()
		select {
		case <-ctx.Done():
			timeoutCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
			server.Shutdown(timeoutCtx)
		}
	}()
	return stopCh, server.Serve(l)
}

// 如果返回的 exitCh 不为空，需要等待收到通知，才表示服务器完成shutdown
func serveDebug(ctx context.Context) (stopCh chan struct{}, err error) {
	addr := "127.0.0.1:8081"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, pkgErr.Wrapf(err, "start [debug server] fail")
	}
	log.Printf("[dubug server] luanch success, running on %s", l.Addr())
	server := &http.Server{Handler: http.DefaultServeMux}

	stopCh = make(chan struct{})

	go func() {
		defer func() {
			// 通知shutdown完成了
			close(stopCh)
		}()
		select {
		case <-ctx.Done():
			timeoutCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
			server.Shutdown(timeoutCtx)
		}
	}()
	return stopCh, server.Serve(l)
}
