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

func Listen(path string, perms int) (*net.UnixListener, error) {
	var oldMask int
	defer func() {
		if perms != 0 {
			syscall.Umask(oldMask)
		}
	}()

	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}

	if perms != 0 {
		oldMask = syscall.Umask(0777 ^ perms)
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

func main() {
	l, err := Listen("./socket", 0)
	if err != nil {
		panic(err)
	}
	l.Accept()
	for {
	}
}
