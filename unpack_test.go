package tarutil

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

const emptyDigest = digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

func TestUntarBadLink(t *testing.T) {
	var (
		pr, pw = io.Pipe()
	)
	go func() {
		tw := tar.NewWriter(pw)
		for i := 0; i < 20; i++ {
			name := fmt.Sprintf("%d", i)

			h := tar.Header{
				Name:     fmt.Sprintf("%s.lnk", name),
				Linkname: name,
				Typeflag: tar.TypeLink,
			}
			tw.WriteHeader(&h)
		}
		tw.Close()
	}()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	err = Unpack(context.Background(), pr, dir, nil)
	if errors.Cause(err) != errInvalidLink {
		t.Fatal("processed invalid hard links")
	}

	pr, pw = io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		for i := 0; i < 20; i++ {
			name := fmt.Sprintf("%d", i)

			h := tar.Header{
				Name:     fmt.Sprintf("%s.lnk", name),
				Linkname: "/etc/passwd",
				Typeflag: tar.TypeLink,
			}
			tw.WriteHeader(&h)
		}
		tw.Close()
	}()

	err = Unpack(context.Background(), pr, dir, nil)
	if errors.Cause(err) != errInvalidLink {
		t.Fatalf("processed invalid hard links: actual error: %v", err)
	}
}

func TestUntarNonExistingDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(dir)

	r := generateTar(25)

	if err := Unpack(context.Background(), r, dir, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUntarExistingDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	r := generateTar(25)

	if err := Unpack(context.Background(), r, dir, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUntarOverFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	r := generateTar(25)

	if err := Unpack(context.Background(), r, f.Name(), nil); err == nil {
		t.Fatal("did not error unpacking over file")
	}
}

func TestUntarTeeReader(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)
	r := generateTar(25)
	dgr := digest.SHA256.Digester()

	tee := io.TeeReader(r, dgr.Hash())

	if err := Unpack(context.Background(), tee, dir, nil); err != nil {
		t.Fatal(err)
	}

	if dgr.Digest() == emptyDigest {
		t.Fatal("digest was empty")
	}
}

func TestUntarReader(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	r := generateTar(25)

	if err := Unpack(context.Background(), r, dir, nil); err != nil {
		t.Fatal(err)
	}
}
