package tarutil

import (
	"archive/tar"
	"fmt"
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
