// Package socket provides helpers for working with Unix domain
// sockets.
package socket

import (
	"net"
	"os"
	"syscall"
)

func isAddressInUse(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		return err.Err.Error() == "bind: address already in use"
	}

	return false
}

func isConnectionRefused(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		return err.Err.Error() == "connection refused"
	}

	return false
}

// Listen attempts to listen on a unix socket given by path. If a
// socket already exists but doesn't accept connections assume that
// it's a stale socket, delete it and try again.
func Listen(path string, perms int) (*net.UnixListener, error) {
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}

	if perms != 0 {
		oldMask := syscall.Umask(0777 ^ perms)
		defer func() {
			syscall.Umask(oldMask)
		}()
	}

	l, firstErr := net.ListenUnix("unix", addr)
	if firstErr != nil {
		if !isAddressInUse(firstErr) {
			return nil, firstErr
		}

		conn, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			if !isConnectionRefused(err) {
				return nil, firstErr
			}

			err := os.Remove(path)
			if err != nil {
				return nil, err // TODO wrap error
			}

			return net.ListenUnix("unix", addr)
		}
		conn.Close()
		return nil, firstErr
	}

	return l, err
}
