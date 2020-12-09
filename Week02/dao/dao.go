package dao

import (
	"errors"
	"github.com/wnate/Go-000/tree/main/Week02/model"
)

var errDataNotFound = errors.New("err: data not found")

var inst = newDaoImpl()

// 是否找不到数据报错
func IsDataNotFoundErr(err error) bool {
	return errors.Is(err, errDataNotFound)
}

type Dao interface {
	GetUser(id int) (*model.User, error)
}

func GetInst() Dao {
	return inst
}
