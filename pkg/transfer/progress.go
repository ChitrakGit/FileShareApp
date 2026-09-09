package transfer

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

// ProgressReader wraps an io.Reader and tracks bytes read.
type ProgressReader struct {
	reader     io.Reader
	totalBytes int64
	readBytes  int64
	startTime  time.Time
	lastTime   time.Time
	lastBytes  int64
	onProgress func(percent float64, speedMBps float64, etaSec int)
}

// NewProgressReader creates a progress reader.
func NewProgressReader(r io.Reader, total int64, onProgress func(percent float64, speedMBps float64, etaSec int)) *ProgressReader {
	now := time.Now()
	return &ProgressReader{
		reader:     r,
		totalBytes: total,
		startTime:  now,
		lastTime:   now,
		onProgress: onProgress,
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
			percent := float64(0)
			if pr.totalBytes > 0 {
				percent = (float64(current) / float64(pr.totalBytes)) * 100.0
			}

			diffBytes := current - pr.lastBytes
			speedBytesSec := float64(diffBytes) / elapsed
			speedMBps := speedBytesSec / (1024 * 1024)

			etaSec := 0
			if speedBytesSec > 0 && pr.totalBytes > current {
				etaSec = int(math.Ceil(float64(pr.totalBytes-current) / speedBytesSec))
			}

			if pr.onProgress != nil {
				pr.onProgress(percent, speedMBps, etaSec)
			}

			pr.lastTime = now
			pr.lastBytes = current
		}
	}
	return n, err
}

// RenderProgressBar prints an in-place animated terminal progress bar.
func RenderProgressBar(percent float64, currentBytes, totalBytes int64, speedMBps float64, etaSec int) {
	width := 30
	completed := int(float64(width) * (percent / 100.0))
	if completed > width {
		completed = width
	}
	if completed < 0 {
		completed = 0
	}

	bar := strings.Repeat("█", completed) + strings.Repeat("░", width-completed)
	fmt.Printf("\r\033[K[%s] %5.1f%% | %s/%s | %.1f MB/s | ETA: %ds",
		bar,
		percent,
		FormatBytes(currentBytes),
		FormatBytes(totalBytes),
		speedMBps,
		etaSec,
	)
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
