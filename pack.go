package tarutil

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

func getLink(source, p string, fi os.FileInfo, inodeTable map[uint64]string) (string, bool, error) {
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		follow, err := filepath.EvalSymlinks(p)
		if err != nil {
			return "", false, errors.Wrap(errInvalidSymlink, err.Error())
		}

		rel, err := filepath.Rel(source, follow)
		if err != nil {
			return "", false, errors.Wrap(errInvalidSymlink, err.Error())
		}

		return rel, false, nil
	}

	inode := fi.Sys().(*syscall.Stat_t).Ino
	nlink := fi.Sys().(*syscall.Stat_t).Nlink

	if nlink > 1 {
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

		header, err := tar.FileInfoHeader(fi, linkName)
		if err != nil {
			return err
		}

		header.Linkname = linkName
		header.Name = rel

		if hardLink {
			header.Typeflag = tar.TypeLink
			header.Size = 0
		} else if linkName != "" {
			header.Typeflag = tar.TypeSymlink
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() && header.Linkname == "" {
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
