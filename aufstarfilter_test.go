package tarutil

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var errHashChainMismatch = "original & new hash don't match"

func TestAUFSWhiteoutRoundTrip(t *testing.T) {
	var (
		h1                    = sha256.New()
		h2                    = sha256.New()
		ovlTarStream, aufsTar io.Reader
		err                   error
	)
	entries, err := loadHeaders("headers.json")
	if err != nil {
		t.Fatalf("encountered error loading heders: %v", err)
	}

	pr, pw := io.Pipe()
	tw := tar.NewWriter(pw)
	go func() {
		for _, v := range entries {
			tw.WriteHeader(&v)
		}
		tw.Close()
		pw.Close()
	}()

	originalTarStream := io.TeeReader(pr, h1)

	ovl := NewOverlayWhiteouts()
	ovlTarStream, err = FilterTarUsingFilter(originalTarStream, ovl)
	if err != nil {
		t.Fatalf("encountered error with filter: %v", err)
	}

	aufs := NewAUFSWhiteouts()
	aufsTar, err = FilterTarUsingFilter(ovlTarStream, aufs)
	if err != nil {
		t.Fatalf("encountered error with filter: %v", err)
	}

	io.Copy(h2, aufsTar)
	originalHash := fmt.Sprintf("%x", h1.Sum(nil))
	filteredHash := fmt.Sprintf("%x", h2.Sum(nil))
	if originalHash != filteredHash {
		t.Fatal(errHashChainMismatch)
	}

}

func loadHeaders(fileName string) ([]tar.Header, error) {
	var headers []tar.Header
	f, err := os.Open(fileName)
	if err != nil {
		err = fmt.Errorf("failed to open file %v: %v", fileName, err)
		return nil, err
	}

	jsonData, err := ioutil.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("failed to read the contents of the file %v: %v", fileName, err)
		return nil, err
	}

	err = json.Unmarshal(jsonData, &headers)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal the data: %v", err)
		return nil, err
	}

	return headers, nil
}
