package tarutil

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

var (
	maxTime                  time.Time
	errSyscallNotImplemented = errors.New("syscall not implemented")
	errFailedOpen            = errors.New("failed to open file")
	errFailedWrite           = errors.New("failed to write file")
	errPathIsNonDirectory    = errors.New("path exists, but it's not a directory")
	errDirectoryExists       = errors.New("expected directory to not exist")
	errDirectoryCreateFailed = errors.New("failed to create directory")
	errInvalidSymlink        = errors.New("invalid symlink")
	errInvalidLink           = errors.New("invalid hard link")
	errRead                  = errors.New("encountered error while reading")
	errUnknownHeader         = errors.New("encountered unknown header")
)

type stringMap map[string]struct{}

// Options controls the behavior of some tarball related operations.
type Options struct {
	NoLchown bool
}

func init() {
	if unsafe.Sizeof(syscall.Timespec{}.Nsec) == 8 {
		// This is a 64 bit timespec
		// os.Chtimes limits time to the following
		maxTime = time.Unix(0, 1<<63-1)
	} else {
		// This is a 32 bit timespec
		maxTime = time.Unix(1<<31-1, 0)
	}
}
