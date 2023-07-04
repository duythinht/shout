package shout

/*
#cgo LDFLAGS: -lshout
#include <stdlib.h>
#include <shout/shout.h>
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

func init() {
	C.shout_init()
}

// ShutDown shuts down the shout library, deallocating any global storage. Don't call
// anything afterwards
func ShutDown() {
	C.shout_shutdown()
}

// Version represents the libshout version struct
type Version struct {
	Version string
	Major   int
	Minor   int
	Patch   int
}

// Error is an libshout integer error type
type Error int

// Error constants
const (
	ShoutErrorSuccess     Error = 0
	ShoutErrorInsane      Error = -1
	ShoutErrorNoConnect   Error = -2
	ShoutErrorNoLogin     Error = -3
	ShoutErrorSocket      Error = -4
	ShoutErrorMalloc      Error = -5
	ShoutErrorMetadata    Error = -6
	ShoutErrorConnected   Error = -7
	ShoutErrorUnconnected Error = -8
	ShoutErrorUnsupported Error = -9
	ShoutErrorBusy        Error = -10
	ShoutErrorNoTLS       Error = -11
	ShoutErrorTLSBadCert  Error = -12
	ShoutErrorRetry       Error = -13
)

func (e Error) Error() string {
	switch e {
	case ShoutErrorSuccess:
		return "success"
	case ShoutErrorInsane:
		return "insane error"
	case ShoutErrorNoConnect:
		return "connect error"
	case ShoutErrorNoLogin:
		return "login error"
	case ShoutErrorSocket:
		return "socket error"
	case ShoutErrorMalloc:
		return "memory allocation error"
	case ShoutErrorMetadata:
		return "metadata error"
	case ShoutErrorConnected:
		return "connected"
	case ShoutErrorUnconnected:
		return "not connected"
	case ShoutErrorUnsupported:
		return "not supported"
	case ShoutErrorBusy:
		return "non-blocking io busy"
	case ShoutErrorNoTLS:
		return "no tls error"
	case ShoutErrorTLSBadCert:
		return "bad certificate"
	case ShoutErrorRetry:
		return "retry"
	default:
		return fmt.Sprintf("unknown error code %d", e)
	}
}

// GetVersion returns libshout version
func GetVersion() *Version {
	var major, minor, patch C.int
	csv := C.shout_version(
		&major,
		&minor,
		&patch,
	)

	result := &Version{
		Version: C.GoString(csv),
		Major:   int(major),
		Minor:   int(minor),
		Patch:   int(patch),
	}

	return result
}

// StreamFormat type represents a stream format
type StreamFormat int

// Format constants
const (
	ShoutFormatOGG       StreamFormat = 0
	ShoutFormatVorbis    StreamFormat = 0
	ShoutFormatMP3       StreamFormat = 1
	ShoutFormatWEBM      StreamFormat = 2
	ShoutFormatWEBMAudio StreamFormat = 3
)

// Protocol type represents a stream protocol type
type Protocol int

// Protocol constants
const (
	ProtocolHTTP       Protocol = 0
	ProtocolXAudioCast Protocol = 1
	ProtocolICY        Protocol = 2
	ProtocolRoarAudio  Protocol = 3
)

func convError(errCode int) error {
	if errCode == int(ShoutErrorSuccess) {
		return nil
	}
	return Error(errCode)
}

// Config is a shout configuration struct
type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Mount    string
	Proto    Protocol
	Format   StreamFormat
}

// Writer represents a Writer interface to libshout streaming structure
type Writer struct {
	cshout *C.struct_shout
}

// Connect creates a connection according to configuration and returns
// a Writer instance
func Connect(cfg *Config) (*Writer, error) {
	var cstr *C.char
	shout := C.shout_new()
	if shout == nil {
		return nil, fmt.Errorf("error allocating shout_t structure")
	}

	// setting hostname
	cstr = C.CString(cfg.Host)
	C.shout_set_host(shout, cstr)
	C.free(unsafe.Pointer(cstr))

	// setting port
	C.shout_set_port(shout, C.ushort(cfg.Port))

	// setting source user
	cstr = C.CString(cfg.User)
	C.shout_set_user(shout, cstr)
	C.free(unsafe.Pointer(cstr))

	// setting source password
	cstr = C.CString(cfg.Password)
	C.shout_set_password(shout, cstr)
	C.free(unsafe.Pointer(cstr))

	// setting mountpoint
	cstr = C.CString(cfg.Mount)
	C.shout_set_mount(shout, cstr)
	C.free(unsafe.Pointer(cstr))

	// setting stream format
	C.shout_set_content_format(shout, C.uint(cfg.Format), C.uint(1), nil)

	// setting stream protocol
	C.shout_set_protocol(shout, C.uint(cfg.Proto))

	res := int(C.shout_open(shout))
	if res != int(ShoutErrorSuccess) {
		C.shout_free(shout)
		err := convError(res)
		return nil, err
	}

	w := &Writer{cshout: shout}
	runtime.SetFinalizer(w, finalize)
	return w, nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	pptr := (*C.uchar)(&p[0])
	bufSize := C.size_t(len(p))
	res := int(C.shout_send(w.cshout, pptr, bufSize))
	err = convError(res)
	if err == nil {
		n = len(p)
		C.shout_sync(w.cshout)
	}
	return
}

// Close closes a shout connection
func (w *Writer) Close() error {
	fmt.Println("shout closed")
	res := int(C.shout_close(w.cshout))
	return convError(res)
}

func (w *Writer) getErrno() int {
	return int(C.shout_get_errno(w.cshout))
}

func finalize(w *Writer) {
	C.shout_free(w.cshout)
}
