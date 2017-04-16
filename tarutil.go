package tarutil

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func luTimesNano(path string, ts []syscall.Timespec) error {
	var (
		_path *byte
		err   error
	)

	// These are not currently available in syscall
	atFdCwd := -100
	atSymLinkNoFollow := 0x100

	if _path, err = syscall.BytePtrFromString(path); err != nil {
		return err
	}

	_, _, res := syscall.Syscall6(syscall.SYS_UTIMENSAT, uintptr(atFdCwd), uintptr(unsafe.Pointer(_path)), uintptr(unsafe.Pointer(&ts[0])), uintptr(atSymLinkNoFollow), 0, 0)
	if res != 0 && res != syscall.ENOSYS {
		return errSyscallNotImplemented
	}

	return nil
}

func chtimes(name string, atime time.Time, mtime time.Time) error {
	unixMinTime := time.Unix(0, 0)
	unixMaxTime := maxTime

	// If the modified time is prior to the Unix Epoch, or after the
	// end of Unix Time, os.Chtimes has undefined behavior
	// default to Unix Epoch in this case, just in case

	if atime.Before(unixMinTime) || atime.After(unixMaxTime) {
		atime = unixMinTime
	}

	if mtime.Before(unixMinTime) || mtime.After(unixMaxTime) {
		mtime = unixMinTime
	}

	return os.Chtimes(name, atime, mtime)
}

func timeToTimespec(time time.Time) syscall.Timespec {
	if time.IsZero() {
		// Return UTIME_OMIT special value
		ts := syscall.Timespec{
			Sec:  0,
			Nsec: ((1 << 30) - 2),
		}
		return ts
	}
	return syscall.NsecToTimespec(time.UnixNano())
}

func directoryExists(dirPath string) (bool, error) {
	fi, err := os.Lstat(dirPath)
	if err == nil && !fi.IsDir() {
		return false, errors.Wrap(errPathIsNonDirectory, dirPath)
	}
	if err != nil {
		return false, nil
	}

	return true, nil
}

func createDirectory(destPath string, fi os.FileInfo) error {
	if _, err := directoryExists(destPath); err != nil {
		return errors.Wrap(errDirectoryExists, destPath)
	}
	if err := os.Mkdir(destPath, fi.Mode()); err != nil {
		return errors.Wrap(errDirectoryCreateFailed, destPath)
	}

	return nil
}

func createFile(destPath string, fi os.FileInfo, r io.Reader) error {
	file, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, fi.Mode())
	if err != nil {
		return errors.Wrap(errFailedOpen, destPath)
	}
	defer file.Close()
	if _, err := io.Copy(file, r); err != nil {
		return errors.Wrap(errFailedWrite, destPath)
	}

	return nil
}

func createSymlink(dest, destPath string, header *tar.Header) error {
	targetPath := filepath.Join(filepath.Dir(destPath), header.Linkname)

	if !strings.HasPrefix(targetPath, dest) {
		return errors.Wrap(errInvalidSymlink, header.Linkname)
	}
	return os.Symlink(header.Linkname, destPath)
}

func mkdev(major, minor int64) uint32 {
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}

func createBlockCharFifo(destPath string, header *tar.Header) error {
	mode := uint32(header.Mode & 07777)
	switch header.Typeflag {
	case tar.TypeBlock:
		mode |= syscall.S_IFBLK
	case tar.TypeChar:
		mode |= syscall.S_IFCHR
	case tar.TypeFifo:
		mode |= syscall.S_IFIFO
	}

	dev := int(mkdev(header.Devmajor, header.Devminor))
	return syscall.Mknod(destPath, mode, dev)
}

func handleWhiteouts(destPath string, unpackedPaths stringMap) error {
	base := filepath.Base(destPath)
	dir := filepath.Dir(destPath)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil // parent was deleted
			}
			return err
		}
		if path == dir {
			return nil
		}
		if _, exists := unpackedPaths[path]; !exists {
			return os.RemoveAll(path)
		}
		return nil
	}

	if base == whiteoutOpaqueDir {
		if _, err := os.Lstat(dir); err != nil {
			return err
		}
		return filepath.Walk(dir, walkFn)
	}

	originalBase := base[len(whiteoutPrefix):]
	originalPath := filepath.Join(dir, originalBase)
	return os.RemoveAll(originalPath)
}

func setPermissions(destPath string, header *tar.Header, options *Options) error {
	if options == nil || !options.NoLchown {
		if err := os.Lchown(destPath, header.Uid, header.Gid); err != nil {
			return err
		}
	}

	headerFi := header.FileInfo()
	if header.Typeflag == tar.TypeLink {
		fi, err := os.Lstat(header.Linkname)
		if err == nil && (fi.Mode()&os.ModeSymlink == 0) {
			return os.Chmod(destPath, headerFi.Mode())
		}
	} else if header.Typeflag != tar.TypeSymlink {
		return os.Chmod(destPath, headerFi.Mode())
	}

	return nil
}

func setMtimeAndAtime(destPath string, header *tar.Header) error {
	aTime := header.AccessTime
	if aTime.Before(header.ModTime) {
		aTime = header.ModTime
	}

	// system.Chtimes doesn't support a NOFOLLOW flag atm
	if header.Typeflag == tar.TypeLink {
		fi, err := os.Lstat(header.Linkname)
		if err == nil && (fi.Mode()&os.ModeSymlink == 0) {
			return chtimes(destPath, aTime, header.ModTime)
		}
	} else if header.Typeflag != tar.TypeSymlink {
		return chtimes(destPath, aTime, header.ModTime)
	} else {
		ts := []syscall.Timespec{timeToTimespec(aTime), timeToTimespec(header.ModTime)}
		return luTimesNano(destPath, ts)
	}
	return nil
}

func handleTarEntry(fullPath, dest string, header *tar.Header, tr io.Reader, options *Options) error {
	var err error
	fi := header.FileInfo()

	switch header.Typeflag {
	case tar.TypeDir:
		err = createDirectory(fullPath, fi)
	case tar.TypeReg, tar.TypeRegA:
		err = createFile(fullPath, fi, tr)
	case tar.TypeBlock, tar.TypeChar, tar.TypeFifo:
		err = createBlockCharFifo(fullPath, header)
	case tar.TypeSymlink:
		err = createSymlink(dest, fullPath, header)
	default:
		err = errors.Wrap(errUnknownHeader, fullPath)
	}
	if err != nil {
		return err
	}

	err = setPermissions(fullPath, header, options)
	if err != nil {
		return err
	}

	return setMtimeAndAtime(fullPath, header)
}

func changeDirTimes(dirs []*tar.Header, dest string) error {
	for _, hdr := range dirs {
		path := filepath.Join(dest, hdr.Name)
		if err := chtimes(path, hdr.AccessTime, hdr.ModTime); err != nil {
			return err
		}
	}
	return nil
}

// UnpackTar unpacks a tar file into the destination.
func UnpackTar(ctx context.Context, r io.Reader, dest string, options *Options) error {
	fi, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dest, 0700); err != nil {
			return errors.Wrap(err, dest)
		}
	} else if os.IsExist(err) && !fi.IsDir() {
		return errors.Wrap(errRead, "destination is not a directory")
	} else if err != nil {
		return errors.Wrap(err, dest)
	}

	tr := tar.NewReader(r)
	unpackedPaths := make(stringMap)

	var dirs []*tar.Header
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return errRead
		}

		fullPath := filepath.Join(dest, hdr.Name)
		base := filepath.Base(fullPath)

		if strings.HasPrefix(base, whiteoutPrefix) {
			handleWhiteouts(fullPath, unpackedPaths)
			continue
		}

		if hdr.Name != "/" {
			if err := handleTarEntry(fullPath, dest, hdr, tr, options); err != nil {
				return err
			}
		}

		if hdr.Typeflag == tar.TypeDir {
			dirs = append(dirs, hdr)
		}
		unpackedPaths[fullPath] = struct{}{}
	}

	return changeDirTimes(dirs, dest)
}

// OpenAndUnpack unpacks a specified file into the destination.
func OpenAndUnpack(ctx context.Context, layerPath, dest string, options *Options) error {
	tarFile, err := os.Open(layerPath)
	if err != nil {
		return errors.Wrap(errFailedOpen, layerPath)
	}
	defer tarFile.Close()

	return UnpackTar(ctx, tarFile, dest, options)
}

// OpenAndUnpackMulti unpacks multiple files into the destination.
func OpenAndUnpackMulti(ctx context.Context, layers []string, dest string, options *Options) error {
	for i := range layers {
		err := OpenAndUnpack(ctx, layers[i], dest, options)
		if err != nil {
			return err
		}
	}

	return nil
}
