package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// 添加关服回调钩子函数
func AddShutDownHook(ctx context.Context, hook func()) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	select {
		case <- ctx.Done():
			hook()
		case <-signalChan:
			hook()
	}
}
