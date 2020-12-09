package biz

var inst = newBizImpl()

type Biz interface {
	// 添加用户金币数
	AddUserGoldCoin(uid int, delta int32) error
}

func GetInst() Biz {
	return inst
}
