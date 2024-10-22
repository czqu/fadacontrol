package bluetooth

/*
#cgo LDFLAGS: -lws2_32
#include "bluetooth_windwos.h"
*/
import "C"
import (
	"errors"
	"net"
	"time"
	"unsafe"
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type BTConn struct {
	socket C.SOCKET
}

type Ble struct {
	socket C.SOCKET
}

func NewBTConn(socket C.SOCKET) *BTConn {
	return &BTConn{socket: socket}
}

type Config struct {
	ServiceInstanceName string
	Comment             string
}

func (l *Ble) Accept() (net.Conn, error) {
	clientSocket := C.Accept(l.socket)
	if int(C.isInvalidSocket(clientSocket)) == 1 {
		return nil, errors.New("failed to accept bluetooth connection")
	}
	return NewBTConn(clientSocket), nil
}

func (l *Ble) Close() error {
	if C.closesocket(l.socket) == C.SOCKET_ERROR {
		return errors.New("failed to close bluetooth socket")
	}
	return nil
}

func (l *Ble) Addr() net.Addr {
	return nil
}

func Listen(serviceClassId GUID, config *Config) (net.Listener, error) {
	cServiceInstanceName := C.CString(config.ServiceInstanceName)
	cComment := C.CString(config.Comment)
	defer C.free(unsafe.Pointer(cServiceInstanceName))
	defer C.free(unsafe.Pointer(cComment))

	cServiceClassId := C.GUID{
		Data1: C.ulong(serviceClassId.Data1),
		Data2: C.ushort(serviceClassId.Data2),
		Data3: C.ushort(serviceClassId.Data3),
		Data4: *(*[8]C.uchar)(unsafe.Pointer(&serviceClassId.Data4)),
	}

	socket := C.Listen(&cServiceClassId, cServiceInstanceName, cComment)
	if int(C.isInvalidSocket(socket)) == 1 {
		return nil, errors.New("failed to listen on bluetooth socket")
	}

	return &Ble{socket: socket}, nil
}

func (bt *BTConn) Read(b []byte) (n int, err error) {
	ret := C.recv(bt.socket, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)), 0)
	if ret == C.SOCKET_ERROR {
		return 0, errors.New("failed to read from bluetooth connection")
	}
	return int(ret), nil
}

func (bt *BTConn) Close() error {
	if C.closesocket(bt.socket) == C.SOCKET_ERROR {
		return errors.New("failed to close bluetooth socket")
	}
	return nil
}

func (bt *BTConn) Write(b []byte) (n int, err error) {
	ret := C.send(bt.socket, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)), 0)
	if ret == C.SOCKET_ERROR {
		return 0, errors.New("failed to write to bluetooth connection")
	}
	return int(ret), nil
}

func (bt *BTConn) LocalAddr() net.Addr {
	return nil
}

func (bt *BTConn) RemoteAddr() net.Addr {
	return nil
}

func (bt *BTConn) SetDeadline(t time.Time) error {
	return bt.setDeadline(t, 'b')
}

func (bt *BTConn) SetReadDeadline(t time.Time) error {
	return bt.setDeadline(t, 'r')
}

func (bt *BTConn) SetWriteDeadline(t time.Time) error {
	return bt.setDeadline(t, 'w')
}
func (bt *BTConn) setDeadline(t time.Time, typ byte) error {
	var tv C.struct_timeval
	if !t.IsZero() {
		dur := t.Sub(time.Now())
		if dur < 0 {
			tv.tv_sec = 0
			tv.tv_usec = 0
		} else {
			tv.tv_sec = C.long(dur / time.Second)
			tv.tv_usec = C.long((dur % time.Second) / time.Microsecond)
		}
	}

	var opt C.int
	switch typ {
	case 'b':
		opt = C.SO_RCVTIMEO | C.SO_SNDTIMEO
	case 'r':
		opt = C.SO_RCVTIMEO
	case 'w':
		opt = C.SO_SNDTIMEO
	}

	ret := C.setsockopt(bt.socket, C.SOL_SOCKET, opt, (*C.char)(unsafe.Pointer(&tv)), C.int(unsafe.Sizeof(tv)))
	if ret == C.SOCKET_ERROR {
		return errors.New("failed to set deadline")
	}
	return nil
}
