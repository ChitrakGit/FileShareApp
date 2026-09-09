package transfer

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamAndExtractTarRoundTrip(t *testing.T) {
	// 1. Create temporary directory structure
	tempSrc := t.TempDir()
	tempDest := t.TempDir()

	file1Path := filepath.Join(tempSrc, "file1.txt")
	subDirPath := filepath.Join(tempSrc, "subdir")
	file2Path := filepath.Join(subDirPath, "file2.txt")

	_ = os.WriteFile(file1Path, []byte("Hello World from File 1"), 0644)
	_ = os.MkdirAll(subDirPath, 0755)
	_ = os.WriteFile(file2Path, []byte("Nested Content in File 2"), 0644)

	// 2. Inspect paths
	items, totalBytes, err := InspectPaths([]string{file1Path, subDirPath})
	if err != nil {
		t.Fatalf("InspectPaths failed: %v", err)
	}

	if len(items) != 3 { // file1, subdir, file2
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	expectedBytes := int64(len("Hello World from File 1") + len("Nested Content in File 2"))
	if totalBytes != expectedBytes {
		t.Fatalf("expected %d total bytes, got %d", expectedBytes, totalBytes)
	}

	// 3. Stream to tar buffer
	var buf bytes.Buffer
	err = StreamTar(items, &buf, nil)
	if err != nil {
		t.Fatalf("StreamTar failed: %v", err)
	}

	// 4. Extract into destination
	count, extractedBytes, err := ExtractTar(&buf, tempDest, nil)
	if err != nil {
		t.Fatalf("ExtractTar failed: %v", err)
	}

	if count != 2 { // 2 regular files
		t.Fatalf("expected 2 files extracted, got %d", count)
	}

	if extractedBytes != expectedBytes {
		t.Fatalf("expected %d extracted bytes, got %d", expectedBytes, extractedBytes)
	}

	// 5. Verify extracted file contents
	extracted1, err := os.ReadFile(filepath.Join(tempDest, "file1.txt"))
	if err != nil || string(extracted1) != "Hello World from File 1" {
		t.Fatalf("file1 content mismatch: %v, %s", err, string(extracted1))
	}

	extracted2, err := os.ReadFile(filepath.Join(tempDest, "subdir", "file2.txt"))
	if err != nil || string(extracted2) != "Nested Content in File 2" {
		t.Fatalf("file2 content mismatch: %v, %s", err, string(extracted2))
	}
}

func TestPathTraversalSecurity(t *testing.T) {
	tempDest := t.TempDir()

	// Handcraft a malicious tar stream with `../evil.txt`
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	header := &tar.Header{
		Name:     "../evil.txt",
		Mode:     0644,
		Size:     4,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	if _, err := tw.Write([]byte("evil")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	tw.Close()

	// ExtractTar must detect and reject the path traversal attempt
	_, _, err := ExtractTar(&buf, tempDest, nil)
	if err == nil {
		t.Fatal("expected security error on path traversal attempt, got nil")
	}
}
