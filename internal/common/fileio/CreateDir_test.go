package fileio

import (
	"os"
	"testing"
)

func TestCreateDirKeepContent(t *testing.T) {
	dir := "test_dir_keep_content"

	// Cleanup after the test
	defer os.RemoveAll(dir)

	// Test creating a new directory
	CreateDirKeepContent(dir)

	// Verify the directory was created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory %s to exist, but got error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", dir)
	}

	// Test re-creating the directory (should not remove existing contents)
	testFile := dir + "/test_file.txt"
	err = os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	CreateDirKeepContent(dir)

	// Verify the test file still exists
	_, err = os.Stat(testFile)
	if err != nil {
		t.Fatalf("expected file %s to exist, but got error: %v", testFile, err)
	}
}

func TestCreateEmptyDir(t *testing.T) {
	dir := "test_dir_empty"

	// Cleanup after the test
	defer os.RemoveAll(dir)

	// Create a directory with some content
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	testFile := dir + "/test_file.txt"
	err = os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test clearing and recreating the directory
	CreateEmptyDir(dir)

	// Verify the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory %s to exist, but got error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", dir)
	}

	// Verify the test file was removed
	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Fatalf("expected file %s to be removed, but it still exists", testFile)
	}
}

func TestCreateDirKeepContentPanic(t *testing.T) {
	invalidDir := "/invalid/path"

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid directory creation, but no panic occurred")
		}
	}()

	// Trigger panic by attempting to create a directory in an invalid path
	CreateDirKeepContent(invalidDir)
}

func TestCreateEmptyDirPanic(t *testing.T) {
	invalidDir := "/invalid/path"

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid directory clearing, but no panic occurred")
		}
	}()

	// Trigger panic by attempting to clear and recreate a directory in an invalid path
	CreateEmptyDir(invalidDir)
}
