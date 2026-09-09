package transfer

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTar reads an uncompressed tar stream and safely extracts all contents into destDir.
func ExtractTar(r io.Reader, destDir string, onProgress func(bytesWritten int64)) (int, int64, error) {
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return 0, 0, err
	}

	if err := os.MkdirAll(absDest, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create destination dir: %w", err)
	}

	tr := tar.NewReader(r)
	fileCount := 0
	var totalBytes int64

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fileCount, totalBytes, fmt.Errorf("tar read error: %w", err)
		}

		// Security: Prevent path traversal attacks
		cleanedRelPath := filepath.Clean(filepath.FromSlash(header.Name))
		if strings.HasPrefix(cleanedRelPath, "..") || filepath.IsAbs(cleanedRelPath) {
			return fileCount, totalBytes, fmt.Errorf("security violation: path traversal detected in '%s'", header.Name)
		}

		targetPath := filepath.Join(absDest, cleanedRelPath)

		// Double check targetPath is strictly inside absDest
		rel, err := filepath.Rel(absDest, targetPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fileCount, totalBytes, fmt.Errorf("security violation: illegal path '%s'", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fileCount, totalBytes, fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Ensure parent folder exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fileCount, totalBytes, fmt.Errorf("failed to create parent dir: %w", err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fileCount, totalBytes, fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			buf := make([]byte, 64*1024)
			for {
				n, readErr := tr.Read(buf)
				if n > 0 {
					if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
						outFile.Close()
						return fileCount, totalBytes, writeErr
					}
					totalBytes += int64(n)
					if onProgress != nil {
						onProgress(int64(n))
					}
				}
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					outFile.Close()
					return fileCount, totalBytes, readErr
				}
			}
			outFile.Close()
			fileCount++
		}
	}

	return fileCount, totalBytes, nil
}
