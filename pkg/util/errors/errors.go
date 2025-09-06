package errors

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotUpdate = errors.New("record not updated")
	ErrImageNotFound   = errors.New("未识别到有新增镜像")
	ErrDisableStatus   = errors.New("状态关闭")
)

func IsNotUpdated(err error) bool {
	return errors.Is(err, ErrRecordNotUpdate)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsImageNotFound(err error) bool {
	return errors.Is(err, ErrImageNotFound)
}

func IsDisableStatus(err error) bool {
	return errors.Is(err, ErrDisableStatus)
}
