package dao

import (
	"errors"
	pkgErr "github.com/pkg/errors"
	"github.com/wnate/Go-000/tree/main/Week02/model"
	"math/rand"
)

var testUser = model.User{
	ID:       123456,
	GoldCoin: 100, // 初始100金币
}

func newDaoImpl() Dao {
	return &daoImpl{}
}

type daoImpl struct {
}

func (d *daoImpl) GetUser(id int) (*model.User, error) {
	if id != testUser.ID {
		// 找不到记录
		return nil, pkgErr.Wrapf(errDataNotFound, "user[%d] not found", id)
	}
	// 假设这里有概率出现网络异常
	if rand.Intn(10) > 2 {
		return nil, pkgErr.Wrapf(errors.New("network err"), "get user[%d] failed, unknown error", id)
	}

	return &testUser, nil
}
