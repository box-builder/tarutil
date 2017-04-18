package tarutil

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

func prepHeader(p, linkName, rel string, hardLink bool, fi os.FileInfo) (*tar.Header, error) {
	header, err := tar.FileInfoHeader(fi, linkName)
	if err != nil {
		return nil, err
	}

	header.Linkname = linkName
	header.Name = rel

	// fixups to special files
	if hardLink {
		header.Typeflag = tar.TypeLink
		header.Size = 0
	} else if linkName != "" {
		header.Typeflag = tar.TypeSymlink
	} else if fi.IsDir() && len(header.Name) > 0 && header.Name[len(header.Name)-1] != '/' {
		header.Name += "/"
	}

	// ripped directly from docker
	capability, _ := Lgetxattr(p, "security.capability")
	if capability != nil {
		header.Xattrs = make(map[string]string)
		header.Xattrs["security.capability"] = string(capability)
	}

	header.ChangeTime = time.Unix(fi.Sys().(*syscall.Stat_t).Ctim.Unix())
	header.AccessTime = time.Unix(fi.Sys().(*syscall.Stat_t).Atim.Unix())
	header.ModTime = time.Unix(fi.Sys().(*syscall.Stat_t).Mtim.Unix())

	return header, nil
}

func getLink(source, p string, fi os.FileInfo, inodeTable map[uint64]string) (string, bool, error) {
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		follow, err := os.Readlink(p)
		if err != nil {
			return "", false, errors.Wrap(errInvalidSymlink, err.Error())
		}

		var rel string

		// coerce the path to be a relative symlink.
		if path.IsAbs(follow) {
			rel, err = filepath.Rel(path.Dir(p), follow)
		} else {
			rel, err = filepath.Rel(path.Dir(p), path.Join(path.Dir(p), follow))
		}

		if err != nil {
			return "", false, errors.Wrap(errInvalidSymlink, err.Error())
		}
		return rel, false, nil
	}

	inode := fi.Sys().(*syscall.Stat_t).Ino
	nlink := fi.Sys().(*syscall.Stat_t).Nlink

	if nlink > 1 {
		// FIXME need to track across devices
		if _, ok := inodeTable[inode]; ok {
			return inodeTable[inode], true, nil
		}

		rel, err := filepath.Rel(source, p)
		if err != nil {
			return "", false, errors.Wrap(errInvalidLink, err.Error())
		}

		inodeTable[inode] = rel
		return "", false, nil
	}

	return "", false, nil
}

// Pack packs a tarball from the specified source, into the writer w. Returns
// an error.
func Pack(ctx context.Context, source string, w io.Writer) error {
	inodeTable := map[uint64]string{}

	tw := tar.NewWriter(w)
	defer tw.Close()

	err := filepath.Walk(source, func(p string, fi os.FileInfo, err error) error {
		if p == source {
			return nil
		}

		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, p)
		if err != nil {
			return err
		}

		linkName, hardLink, err := getLink(source, p, fi, inodeTable)
		if err != nil {
			return err
		}

		header, err := prepHeader(p, linkName, rel, hardLink, fi)
		if err != nil {
			return err
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() && (header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA) {
			abs, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			f, err := os.Open(abs)
			if err != nil {
				return err
			}
			_, err = io.Copy(tw, f)
			if err != nil {
				return err
			}
			f.Close()
		}
		return nil
	})

	return err
}
