package tarutil

import "testing"

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
