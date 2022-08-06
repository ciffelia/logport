//go:build !windows

package connection

import (
	"errors"
	"net"
)

func IsClosedError(err error) bool {
	return errors.Is(err, net.ErrClosed)
}
