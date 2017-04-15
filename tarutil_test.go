package tarutil

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	_ "crypto/sha256"

	digest "github.com/opencontainers/go-digest"
)

const emptyDigest = digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

func TestUntarTeeReader(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)
	r := generateTar(25)
	dgr := digest.SHA256.Digester()

	tee := io.TeeReader(r, dgr.Hash())

	if err := UnpackTar(context.Background(), tee, dir); err != nil {
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

	if err := UnpackTar(context.Background(), r, dir); err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryExists(t *testing.T) {
	ok, err := directoryExists("/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Fatalf("/tmp doesn't exist")
	}

	ok, err = directoryExists("/NON/Existing/Directory")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ok {
		t.Fatalf("the directory shouldn't exist")
	}

	ok, err = directoryExists("/dev/null")
	if err != errPathIsNonDirectory {
		t.Fatalf("expected error: %v", err)
	}

	if ok {
		t.Fatalf("DirectoryExists thinks file is a directory")
	}
}
