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

func TestConsumeTar(t *testing.T) {
	r := generateTar(25)
	items, err := loopTar(r, true)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

	if items != 25*2 {
		t.Fatalf("processed only %v items, not 25", items)
	}
}

func TestTarFilterWithNullFilter(t *testing.T) {
	numItems := 25
	r := generateTar(numItems)
	// we also need to read the links from the generated tarball
	numItems = numItems * 2
	nf := &NullFilter{}
	fr, err := FilterTarUsingFilter(r, nf)
	if err != nil {
		t.Fatalf("failed to instantiate filter")
	}

	items, err := loopTar(fr, true)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

	if items != numItems {
		t.Fatalf("processed only %v items, not %v", items, numItems)
	}
}

func TestOverlayWhiteoutsWithDummyFiles(t *testing.T) {
	numItems := 25
	r := generateTar(numItems)
	// we also need to read the links from the generated tarball
	numItems = numItems * 2
	nf := NewOverlayWhiteouts()
	fr, err := FilterTarUsingFilter(r, nf)
	if err != nil {
		t.Fatalf("failed to instantiate filter")
	}

	items, err := loopTar(fr, false)
	if err != nil {
		t.Fatalf("encountered error: %v", err)
	}

	if items != numItems {
		t.Fatalf("processed only %v items, not %v", items, numItems)
	}
}
