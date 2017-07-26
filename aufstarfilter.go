package tarutil

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"
)

// AUFSWhiteouts is a TarFilter to convert overlay whiteouts to aufs whiteouts.
type AUFSWhiteouts struct {
	previousEntry *tar.Header
	tw            *tar.Writer
}

// NewAUFSWhiteouts creates a new overlay whiteout filter.
func NewAUFSWhiteouts() *AUFSWhiteouts {
	return &AUFSWhiteouts{}

}

// SetTarWriter sets the tar writer for output processing.
func (o *AUFSWhiteouts) SetTarWriter(tw *tar.Writer) error {
	if o.tw == nil {
		o.tw = tw
		return nil
	}
	return fmt.Errorf("the TarWriter is already set")
}

// Close closes the tar filter, finalizing any processing.
func (o *AUFSWhiteouts) Close() error {
	return o.tw.Close()
}

// HandleEntry is the meat and potatoes of the filter; managing the overlay files.
func (o *AUFSWhiteouts) HandleEntry(h *tar.Header) (bool, bool, error) {
	fi := h.FileInfo()
	if fi.Mode()&os.ModeCharDevice != 0 && h.Devmajor == 0 && h.Devminor == 0 {
		dir, filename := filepath.Split(h.Name)
		h.Name = filepath.Join(dir, whiteoutPrefix+filename)
		h.Typeflag = tar.TypeReg
		h.Size = 0
		return false, true, nil
	}

	if fi.Mode()&os.ModeDir != 0 {
		opaque, ok := h.Xattrs[overlayOpaqueXattr]
		if ok && opaque == overlayOpaqueXattrValue {
			delete(h.Xattrs, overlayOpaqueXattr)
			h.Typeflag = tar.TypeReg
			h.Name = filepath.Join(h.Name, whiteoutOpaqueDir)
			h.Size = 0
		}
		return false, true, nil
	}

	return true, true, nil
}
