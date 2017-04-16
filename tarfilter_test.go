package tarutil

import (
	"archive/tar"
	"fmt"
	"io"
	"testing"
)

type NullFilter struct {
	tw *tar.Writer
}

func (n *NullFilter) SetTarWriter(tw *tar.Writer) error {
	if n.tw == nil {
		n.tw = tw
		return nil
	}
	return fmt.Errorf("the TarWriter is already set")
}

func (n *NullFilter) Close() error {
	return nil
}

func (n *NullFilter) HandleEntry(h *tar.Header) (bool, bool, error) {
	return true, true, nil
}

func generateTar(numEntries int) io.Reader {
	var (
		pr, pw = io.Pipe()
		tw     = tar.NewWriter(pw)
	)
	go func() {
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

func TestTarFilterWithNullFilter(t *testing.T) {
	r := generateTar(25)
	nf := &NullFilter{}
	fr, err := FilterTarUsingFilter(r, nf)
	tr := tar.NewReader(fr)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

	err = loopTar(tr, false)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}
}

func TestOverlayWhiteoutsWithDummyFiles(t *testing.T) {
	r := generateTar(25)
	nf := NewOverlayWhiteouts()
	fr, err := FilterTarUsingFilter(r, nf)
	tr := tar.NewReader(fr)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

	err = loopTar(tr, false)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

}
