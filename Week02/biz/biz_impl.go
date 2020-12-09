package biz

import (
	"fmt"
	"github.com/wnate/Go-000/tree/main/Week02/dao"
	"sync/atomic"
)
import pkgErr "github.com/pkg/errors"

func newBizImpl() Biz {
	return &bizImpl{}
}

type bizImpl struct {
}

func (b *bizImpl) AddUserGoldCoin(uid int, delta int32) error {
	user, err := dao.GetInst().GetUser(uid)
	if err != nil && dao.IsDataNotFoundErr(err) {
		// 这里可以根据实际业务，看情况要不要特殊处理
	}
	if err != nil {
		return pkgErr.WithMessage(err, fmt.Sprintf("user[%d] not found", uid))
	}

	var curVal, newVal int32
	for ; ; {
		curVal = user.GoldCoin
		newVal = curVal + delta

		if newVal < 0 {
			// min limit
			return pkgErr.New(fmt.Sprintf("user[%d] have no enough gold coin, current val is %d", uid, curVal))
		}

		if newVal > 10000 {
			// max limit
			newVal = 10000
		}

		if atomic.CompareAndSwapInt32(&user.GoldCoin, curVal, newVal) {
			break
		}
	}

	return nil
}
