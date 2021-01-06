package main

import (
	"container/ring"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	lt := NewLimiter(10, 100)
	lt.Start()

	var totalReqCount = 0
	var totalAllowCount = 0

	go func() {
		// 模拟随机请求
		for {
			select {
			case <-time.After(1 * time.Second):
				reqCount := rand.Intn(20)
				totalReqCount += reqCount
				for ; reqCount > 0; reqCount-- {
					if lt.IsAllow() {
						totalAllowCount++
					}
				}
			}
		}

	}()

	for {
		// 定时打印统计数据
		select {
		case <-time.After(5 * time.Second):
			log.Printf("current count=%d, totalReqCount=%d, totalAllowCount=%d\n", lt.curCount, totalReqCount, totalAllowCount)
		}
	}
}

func NewLimiter(limitSeconds int, limitCount int32) *limiter {
	lt := &limiter{
		limitCount: limitCount,
		stopCh:     make(chan struct{}),
	}
	var numBuckets = limitSeconds
	lt.curBucket = ring.New(numBuckets)

	for i := 0; i < numBuckets; i++ {
		lt.curBucket.Value = &counter{}
		lt.curBucket = lt.curBucket.Next()
	}
	return lt
}

type limiter struct {
	limitCount int32
	curCount   int32
	curBucket  *ring.Ring
	stopCh     chan struct{}
	stopOnce   sync.Once
}

type counter struct {
	val int32
}

func (l *limiter) Start() {
	go func() {
		for {
			select {
			case <-l.stopCh:
				return
			case <-time.After(1 * time.Second):
				// 定时每秒滑动一个窗口(桶)，清空并回收下一个桶(即最旧的一个桶的数据)
				nextBucket := l.curBucket.Next()
				nextBucketCount := l.curBucket.Next().Value.(*counter).val
				atomic.AddInt32(&l.curCount, -nextBucketCount)
				nextBucket.Value = &counter{}
				l.curBucket = nextBucket
			}
		}

	}()
}

func (l *limiter) IsAllow() bool {
	if l.curCount >= l.limitCount {
		return false
	}
	var newCount = atomic.AddInt32(&l.curCount, 1)
	if newCount > l.limitCount {
		// 回滚操作
		atomic.AddInt32(&l.limitCount, -1)
		return false
	}
	// 累加当前桶的次数
	bucketCount := l.curBucket.Value.(*counter)
	newCount = atomic.AddInt32(&bucketCount.val, 1)

	return true
}

func (l *limiter) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
}
