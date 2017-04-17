package tarutil

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func generateWord(length int) string {
	a := int('a')
	z := int('z')
	str := ""

	for l := 0; l < length; l++ {
		pos := rand.Intn(z - a)
		str += string(rune(a + pos))
	}

	return str
}

func generateFiles(num, filelen int) (string, []string, error) {
	basePath, err := ioutil.TempDir("", "")
	if err != nil {
		return "", nil, err
	}

	random, err := os.Open("/dev/urandom")
	if err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(basePath, 0700); err != nil {
		return "", nil, err
	}

	ret := []string{}

	for i := 0; i < num; i++ {
		fn := generateWord(filelen)
		fullPath := filepath.Join(basePath, fn)
		f, err := os.Create(fullPath)
		if err != nil {
			return "", nil, err
		}

		if _, err := io.Copy(f, io.LimitReader(random, rand.Int63n(1000000))); err != nil {
			return "", nil, err
		}

		f.Close()

		ret = append(ret, f.Name())
	}

	return basePath, ret, nil
}

func TestPack(t *testing.T) {
	packDir, files, err := generateFiles(10, 15)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(packDir)

	r, w := io.Pipe()

	//tr := tar.NewReader(r)
	go func() {
		w.CloseWithError(Pack(context.Background(), packDir, w))
	}()

	unpackDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(unpackDir)

	if err := UnpackTar(context.Background(), r, unpackDir, nil); err != nil {
		t.Fatal(err)
	}

	var abort string

	err = filepath.Walk(unpackDir, func(p string, fi os.FileInfo, err error) error {
		abort = p

		for _, file := range files {
			if path.Base(file) == path.Base(p) {
				abort = ""
			}
		}

		if len(abort) != 0 {
			return nil
		}

		return nil
	})

	if err != nil || len(abort) != 0 {
		t.Fatalf("Aborted due to missing file %q: error (if any): %v", abort, err)
	}
}
