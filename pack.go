package tarutil

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
)

// Pack packs a tarball from the specified source, into the writer w. Returns
// an error.
func Pack(ctx context.Context, source string, w io.Writer) error {
	tw := tar.NewWriter(w)
	err := filepath.Walk(source, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, p)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		header.Name = rel

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
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
			return f.Close()
		}

		return nil
	})

	tw.Close()
	return err
}
