package transfer

import (
	"fmt"
	"io"
	"math"
	"sync/atomic"
	"time"
)

// ProgressReader wraps an io.Reader and tracks bytes read.
type ProgressReader struct {
	reader     io.Reader
	totalBytes int64
	totalFiles int64
	readBytes  int64
	startTime  time.Time
	lastTime   time.Time
	lastBytes  int64
	onProgress func(speedMBps float64, etaSec int, sentFiles, totalFiles int64, currentBytes, totalBytes int64)
	filesSent  func() int64
}

// NewProgressReader creates a progress reader.
func NewProgressReader(r io.Reader, totalBytes, totalFiles int64, onProgress func(speedMBps float64, etaSec int, sentFiles, totalFiles int64, currentBytes, totalBytes int64), filesSent func() int64) *ProgressReader {
	now := time.Now()
	return &ProgressReader{
		reader:     r,
		totalBytes: totalBytes,
		totalFiles: totalFiles,
		startTime:  now,
		lastTime:   now,
		onProgress: onProgress,
		filesSent:  filesSent,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.readBytes, int64(n))
		now := time.Now()
		elapsed := now.Sub(pr.lastTime).Seconds()

		if elapsed >= 0.25 || err == io.EOF {
			current := atomic.LoadInt64(&pr.readBytes)

			diffBytes := current - pr.lastBytes
			speedBytesSec := float64(diffBytes) / elapsed
			speedMBps := speedBytesSec / (1024 * 1024)

			etaSec := 0
			if speedBytesSec > 0 && pr.totalBytes > current {
				etaSec = int(math.Ceil(float64(pr.totalBytes-current) / speedBytesSec))
			}

			sent := int64(0)
			if pr.filesSent != nil {
				sent = pr.filesSent()
			}

			if pr.onProgress != nil {
				pr.onProgress(speedMBps, etaSec, sent, pr.totalFiles, current, pr.totalBytes)
			}

			pr.lastTime = now
			pr.lastBytes = current
		}
	}
	return n, err
}

// FormatBytes formats byte count into a readable human string (KB, MB, GB).
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
