package main

import (
	"fmt"
	"github.com/wnate/Go-000/tree/main/Week02/biz"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var err error

	if rand.Intn(10) > 5 {
		// 查不到用户
		err = biz.GetInst().AddUserGoldCoin(1234567, -200)
	} else {
		// dao网络异常 或者 用户金币不足
		err = biz.GetInst().AddUserGoldCoin(123456, -200)
	}

	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
