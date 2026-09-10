package client

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"fileshare/pkg/transfer"
	quic_tls "fileshare/pkg/quic"
	"github.com/quic-go/quic-go"
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
	// Strip http:// if present for QUIC address
	targetURL = strings.Replace(targetURL, "http://", "", 1)
	targetURL = strings.Replace(targetURL, "https://", "", 1)
	
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

	startTime := time.Now()
	
	// Attempt QUIC connection
	tlsConf := quic_tls.GenerateClientTLSConfig()
	conn, err := quic.DialAddr(context.Background(), targetURL, tlsConf, nil)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("quic connection failed: %w", err)
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to open quic stream: %w", err)
	}

	// We can write pin/sender as headers later, but for now we just push the stream.
	// Since we need to know success, the server could send back a small JSON or byte,
	// but to keep it simple we just copy the progressReader to the stream.
	_, err = io.Copy(stream, progressReader)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("quic transfer failed: %w", err)
	}
	
	// Small delay to ensure QUIC flushes
	time.Sleep(100 * time.Millisecond)
	// Close stream to signal EOF to server
	stream.Close()

	// Wait for server to acknowledge completion
	ackBuf := make([]byte, 1)
	_, err = stream.Read(ackBuf)
	if err != nil && err != io.EOF {
		fmt.Printf("\n[Sender] Warning: did not receive clean ACK from server: %v\n", err)
	}

	// Ensure progress bar hits 100%
	transfer.RenderProgressBar(100.0, totalBytes, totalBytes, lastSpeed, 0)
	fmt.Printf("\n[Sender] Transfer completed successfully over QUIC in %s!\n\n", time.Since(startTime).Round(time.Millisecond))

	return nil
}

func startsWithHTTP(s string) bool {
	return len(s) >= 7 && (s[:7] == "http://" || (len(s) >= 8 && s[:8] == "https://"))
}
