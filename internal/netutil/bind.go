package netutil

import (
	"context"
	"errors"
	"net"
	"runtime"
	"syscall"
)

func ListenWithOptionalDevice(ctx context.Context, network, address, device string) (net.Listener, error) {
	lc := net.ListenConfig{}
	if device != "" && runtime.GOOS == "linux" {
		lc.Control = func(_, _ string, c syscall.RawConn) error {
			var controlErr error
			if err := c.Control(func(fd uintptr) {
				controlErr = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, device)
			}); err != nil {
				return err
			}
			return controlErr
		}
	}

	ln, err := lc.Listen(ctx, network, address)
	if err != nil && device != "" && runtime.GOOS != "linux" {
		return nil, errors.New("device binding is only supported on linux")
	}
	return ln, err
}
