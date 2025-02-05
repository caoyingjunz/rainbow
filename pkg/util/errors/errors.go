package errors

import "errors"

var (
	ErrRecordNotUpdate = errors.New("record not updated")
)

func IsNotUpdated(err error) bool {
	return errors.Is(err, ErrRecordNotUpdate)
}
