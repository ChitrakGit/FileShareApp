package transfer

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ItemInfo stores metadata of a file or directory to transfer.
type ItemInfo struct {
	SourcePath   string
	RelativePath string
	Size         int64
	IsDir        bool
}

// InspectPaths inspects a list of file/folder paths and computes total size and file list.
func InspectPaths(paths []string) ([]ItemInfo, int64, error) {
	var items []ItemInfo
	var totalSize int64

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid path %s: %w", p, err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot access %s: %w", p, err)
		}

		if info.IsDir() {
			baseDir := filepath.Dir(absPath)
			err := filepath.Walk(absPath, func(currPath string, currInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				relPath, err := filepath.Rel(baseDir, currPath)
				if err != nil {
					return err
				}
				relPath = filepath.ToSlash(relPath)

				item := ItemInfo{
					SourcePath:   currPath,
					RelativePath: relPath,
					Size:         currInfo.Size(),
					IsDir:        currInfo.IsDir(),
				}
				items = append(items, item)
				if !currInfo.IsDir() {
					totalSize += currInfo.Size()
				}
				return nil
			})
			if err != nil {
				return nil, 0, err
			}
		} else {
			items = append(items, ItemInfo{
				SourcePath:   absPath,
				RelativePath: filepath.Base(absPath),
				Size:         info.Size(),
				IsDir:        false,
			})
			totalSize += info.Size()
		}
	}

	return items, totalSize, nil
}

// StreamTar streams a list of ItemInfo entries into a writer using archive/tar on the fly.
func StreamTar(items []ItemInfo, w io.Writer, onProgress func(bytesWritten int64), onFileDone func()) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	checksums := make(map[string]string)

	for _, item := range items {
		info, err := os.Lstat(item.SourcePath)
		if err != nil {
			return fmt.Errorf("stat error for %s: %w", item.SourcePath, err)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("tar header error for %s: %w", item.SourcePath, err)
		}

		// Use normalized relative path inside tar
		header.Name = item.RelativePath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header error: %w", err)
		}

		if !item.IsDir {
			file, err := os.Open(item.SourcePath)
			if err != nil {
				return fmt.Errorf("open file error %s: %w", item.SourcePath, err)
			}

			// Copy with progress and hash
			hasher := sha256.New()
			multiWriter := io.MultiWriter(tw, hasher)
			buf := make([]byte, 64*1024)
			for {
				n, readErr := file.Read(buf)
				if n > 0 {
					if _, writeErr := multiWriter.Write(buf[:n]); writeErr != nil {
						file.Close()
						return writeErr
					}
					if onProgress != nil {
						onProgress(int64(n))
					}
				}
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					file.Close()
					return readErr
				}
			}
			file.Close()
			checksums[item.RelativePath] = fmt.Sprintf("%x", hasher.Sum(nil))
			if onFileDone != nil {
				onFileDone()
			}
		} else {
			if onFileDone != nil {
				onFileDone()
			}
		}
	}

	// Write checksums as the last file in the tar
	checksumData, err := json.Marshal(checksums)
	if err == nil {
		checksumHeader := &tar.Header{
			Name: ".fileshare_checksums.json",
			Mode: 0644,
			Size: int64(len(checksumData)),
		}
		if err := tw.WriteHeader(checksumHeader); err == nil {
			tw.Write(checksumData)
		}
	}

	return nil
}
