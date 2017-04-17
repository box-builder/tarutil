package tarutil

import (
	"archive/tar"
	"fmt"
	"io"
)

func generateTar(numEntries int) io.Reader {
	var (
		pr, pw = io.Pipe()
	)
	go func() {
		tw := tar.NewWriter(pw)
		for i := 0; i < numEntries; i++ {
			name := fmt.Sprintf("foo%v", i)
			h := tar.Header{
				Name:     name,
				Size:     0,
				Uid:      0,
				Gid:      0,
				Typeflag: tar.TypeReg,
			}
			tw.WriteHeader(&h)

			h = tar.Header{
				Name:     fmt.Sprintf("%s.lnk", name),
				Linkname: name,
				Typeflag: tar.TypeLink,
			}
			tw.WriteHeader(&h)
		}
		tw.Close()
	}()

	return pr
}

func loopTar(r io.Reader, print bool) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		if print {
			fmt.Printf("hdr: %#v\n", hdr)
		}
	}
	return nil
}
