package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"fileshare/pkg/transfer"
)

// SendPaths inspects single/multiple files or directories, streams them as tar, and prints live progress.
func SendPaths(targetAddr string, paths []string, pin string, senderName string) error {
	items, totalBytes, err := transfer.InspectPaths(paths)
	if err != nil {
		return fmt.Errorf("failed inspecting paths: %w", err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no files found to send")
	}

	fmt.Printf("\n[Sender] Preparing to send %d items (%s) to %s...\n",
		len(items), transfer.FormatBytes(totalBytes), targetAddr)

	targetURL := targetAddr
	if !startsWithHTTP(targetURL) {
		targetURL = "http://" + targetURL
	}
	targetURL = targetURL + "/api/stream-upload"

	// Create pipe for zero-disk-overhead streaming
	pipeReader, pipeWriter := io.Pipe()

	// Goroutine to pack items into tar stream on the fly
	go func() {
		err := transfer.StreamTar(items, pipeWriter, nil)
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
		} else {
			_ = pipeWriter.Close()
		}
	}()

	// Wrap pipeReader with progress tracking
	var lastSpeed float64

	progressReader := transfer.NewProgressReader(pipeReader, totalBytes, func(percent float64, speedMBps float64, etaSec int) {
		lastSpeed = speedMBps
		current := int64(float64(totalBytes) * (percent / 100.0))
		transfer.RenderProgressBar(percent, current, totalBytes, speedMBps, etaSec)
	})

	req, err := http.NewRequest(http.MethodPost, targetURL, progressReader)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-tar")
	if pin != "" {
		req.Header.Set("X-Fileshare-Pin", pin)
	}
	if senderName != "" {
		req.Header.Set("X-Fileshare-Sender", senderName)
	}

	httpClient := &http.Client{
		Timeout: 0, // No timeout for large file transfers
	}

	startTime := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("transfer failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("transfer rejected by peer (status %d): %s", resp.StatusCode, string(body))
	}

	// Ensure progress bar hits 100%
	transfer.RenderProgressBar(100.0, totalBytes, totalBytes, lastSpeed, 0)
	fmt.Printf("\n[Sender] Transfer completed successfully in %s!\n\n", time.Since(startTime).Round(time.Millisecond))

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if count, ok := result["filesCount"].(float64); ok {
			fmt.Printf("Peer saved %d items.\n", int(count))
		}
	}

	return nil
}

func startsWithHTTP(s string) bool {
	return len(s) >= 7 && (s[:7] == "http://" || (len(s) >= 8 && s[:8] == "https://"))
}
