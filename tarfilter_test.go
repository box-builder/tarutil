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

func TestConsumeTar(t *testing.T) {
	r := generateTar(25)
	items, err := loopTar(r, false)
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

	items, err := loopTar(fr, false)
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

func TestOverlayWhiteouts(t *testing.T) {
	var (
		pr, pw = io.Pipe()
	)
	go func() {
		var items = []struct {
			name      string
			entryType byte
		}{
			{"emptydir", tar.TypeDir},
			{"foo", tar.TypeReg},
			{"bar", tar.TypeDir},
			{fmt.Sprintf("bar/%v", whiteoutOpaqueDir), tar.TypeReg},
			{"boo", tar.TypeDir},
			{fmt.Sprintf("boo/%vbaz", whiteoutPrefix), tar.TypeReg},
			{"lastemptydir", tar.TypeDir},
		}
		tw := tar.NewWriter(pw)

		for _, item := range items {
			h := tar.Header{
				Name:     item.name,
				Size:     0,
				Uid:      0,
				Gid:      0,
				Typeflag: item.entryType,
			}
			tw.WriteHeader(&h)
		}

		tw.Close()
	}()

	filter := NewOverlayWhiteouts()
	filteredTar, err := FilterTarUsingFilter(pr, filter)
	if err != nil {
		t.Fatalf("failed to add filter %v", err)
	}

	headers, err := loopTarAndReturnHeaders(filteredTar)
	if err != nil {
		t.Fatalf("failed to iterate through the items: %v", err)
	}

	var items = []struct {
		name      string
		entryType byte
	}{
		{"emptydir", tar.TypeDir},
		{"foo", tar.TypeReg},
		{"bar", tar.TypeDir},
		{"boo", tar.TypeDir},
		{"boo/baz", tar.TypeChar},
		{"lastemptydir", tar.TypeDir},
	}

	if len(headers) != len(items) {
		t.Fatalf("expected %v headers, got %v", len(items), len(headers))
	}

	for i := range items {
		item := items[i]
		header := headers[i]
		if header.Name != item.name {
			t.Fatalf("expected header.Name: %v, got: %v", item.name, header.Name)
		}
		if header.Typeflag != item.entryType {
			t.Fatalf("expected header.Name: %v, got: %v", item.name, header.Name)
		}
		if header.Name != "bar" {
			continue
		}

		if header.Xattrs["trusted.overlay.opaque"] != "y" {
			t.Fatal("expected entry bar to have overlay xattrs")
		}
	}
}
