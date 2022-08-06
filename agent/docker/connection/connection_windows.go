//go:build windows

package connection

import (
	"errors"
	"github.com/Microsoft/go-winio"
)

func IsClosedError(err error) bool {
	return errors.Is(err, winio.ErrFileClosed)
}
