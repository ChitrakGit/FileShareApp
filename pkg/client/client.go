package client

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"fileshare/pkg/transfer"
	quic_tls "fileshare/pkg/quic"
	"github.com/quic-go/quic-go"
)

// SendPaths inspects single/multiple files or directories, streams them as tar, and prints live progress.
func SendPaths(targetAddrs []string, paths []string, pin string, senderName string) error {
	items, totalBytes, err := transfer.InspectPaths(paths)
	if err != nil {
		return fmt.Errorf("failed inspecting paths: %w", err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no files found to send")
	}

	fmt.Printf("\n[Sender] Preparing to send %d items (%s) to %d recipients...\n",
		len(items), transfer.FormatBytes(totalBytes), len(targetAddrs))

	// Create a custom multi-broadcaster to handle dropping slow receivers
	broadcaster := newBroadcaster(targetAddrs, pin, senderName)
	err = broadcaster.connectAll()
	if err != nil {
		return fmt.Errorf("failed to connect to recipients: %w", err)
	}
	defer broadcaster.closeAll()

	// Create pipe for zero-disk-overhead streaming
	pipeReader, pipeWriter := io.Pipe()

	// Goroutine to pack items into tar stream on the fly
	go func() {
		err := transfer.StreamTar(items, pipeWriter, nil, broadcaster.onProgressFileSent)
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
		} else {
			_ = pipeWriter.Close()
		}
	}()

	// Wrap pipeReader with progress tracking
	progressReader := transfer.NewProgressReader(pipeReader, totalBytes, int64(len(items)), broadcaster.onProgress, func() int64 {
		broadcaster.mu.Lock()
		defer broadcaster.mu.Unlock()
		return broadcaster.filesSent
	})

	startTime := time.Now()

	// We can write pin/sender as headers later, but for now we just push the stream.
	// Since we need to know success, the server could send back a small JSON or byte,
	// but to keep it simple we just copy the progressReader to the stream.
	_, err = io.Copy(broadcaster, progressReader)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("quic transfer failed: %w", err)
	}

	// Wait for streams to close and ACKs to be received
	broadcaster.finalize()

	// Ensure progress bar hits 100%
	broadcaster.renderFinalProgress(totalBytes)
	fmt.Printf("\n[Sender] Transfer completed successfully over QUIC in %s!\n\n", time.Since(startTime).Round(time.Millisecond))

	return nil
}

type broadcaster struct {
	targets       []string
	conns         []*quic.Conn
	streams       []*quic.Stream
	activeStreams []bool
	mu            sync.Mutex
	pin           string
	senderName    string
	filesSent     int64
	lastSpeed     float64
	etaSec        int
	
	// Timeout for writing to a slow receiver
	writeTimeout  time.Duration
}

func newBroadcaster(targets []string, pin, senderName string) *broadcaster {
	return &broadcaster{
		targets:       targets,
		pin:           pin,
		senderName:    senderName,
		writeTimeout:  10 * time.Second, // Drop connection if blocked for > 10 seconds
		filesSent:     0,
	}
}

func (b *broadcaster) connectAll() error {
	tlsConf := quic_tls.GenerateClientTLSConfig()
	
	for _, targetAddr := range b.targets {
		targetURL := targetAddr
		targetURL = strings.Replace(targetURL, "http://", "", 1)
		targetURL = strings.Replace(targetURL, "https://", "", 1)
		
		conn, err := quic.DialAddr(context.Background(), targetURL, tlsConf, nil)
		if err != nil {
			fmt.Printf("\n[Sender] Warning: quic connection failed to %s: %v\n", targetURL, err)
			continue
		}
		
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			conn.CloseWithError(0, "")
			fmt.Printf("\n[Sender] Warning: failed to open quic stream to %s: %v\n", targetURL, err)
			continue
		}
		
		b.conns = append(b.conns, conn)
		b.streams = append(b.streams, stream)
		b.activeStreams = append(b.activeStreams, true)
	}
	
	if len(b.conns) == 0 {
		return fmt.Errorf("could not connect to any recipients")
	}
	
	return nil
}

func (b *broadcaster) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	activeCount := 0
	
	var wg sync.WaitGroup
	for i := range b.streams {
		if !b.activeStreams[i] {
			continue
		}
		activeCount++
		
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Use a channel to enforce write timeout
			done := make(chan error, 1)
			go func() {
				// To keep simple, we just write to the stream. Quic streams have built-in flow control.
				// If a receiver is slow, the stream's window fills up, and Write blocks.
				_, writeErr := b.streams[idx].Write(p)
				done <- writeErr
			}()
			
			select {
			case err := <-done:
				if err != nil {
					b.dropReceiver(idx, fmt.Sprintf("write error: %v", err))
				}
			case <-time.After(b.writeTimeout):
				b.dropReceiver(idx, "write timeout (slow receiver)")
			}
		}(i)
	}
	
	wg.Wait()
	
	if activeCount == 0 {
		return 0, fmt.Errorf("all recipients have been dropped")
	}
	
	// We return len(p) and nil error if at least one receiver is still active,
	// so the io.Copy doesn't stop.
	return len(p), nil
}

func (b *broadcaster) dropReceiver(idx int, reason string) {
	if b.activeStreams[idx] {
		fmt.Printf("\n[Sender] Dropped receiver %s: %s\n", b.targets[idx], reason)
		b.activeStreams[idx] = false
		b.streams[idx].Close()
		b.conns[idx].CloseWithError(0, "")
	}
}

func (b *broadcaster) finalize() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	var wg sync.WaitGroup
	for i := range b.streams {
		if !b.activeStreams[i] {
			continue
		}
		
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Close stream to signal EOF to server
			b.streams[idx].Close()
			
			// Wait for server to acknowledge completion
			ackBuf := make([]byte, 1)
			// Small timeout for ACK
			b.streams[idx].SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err := b.streams[idx].Read(ackBuf)
			if err != nil && err != io.EOF {
				// Just a warning
			}
		}(i)
	}
	wg.Wait()
}

func (b *broadcaster) closeAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for i := range b.conns {
		if b.activeStreams[i] {
			b.streams[i].Close()
			b.conns[i].CloseWithError(0, "")
			b.activeStreams[i] = false
		}
	}
}

func (b *broadcaster) onProgressFileSent() {
	b.mu.Lock()
	b.filesSent++
	b.mu.Unlock()
}

func (b *broadcaster) onProgress(speedMBps float64, etaSec int, sentFiles, totalFiles int64, currentBytes, totalBytes int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.lastSpeed = speedMBps
	b.etaSec = etaSec
	
	// Only speed size/sec and pending time + send files count/ all files count
	fmt.Printf("\r\033[KSpeed: %.1f MB/s | Pending: %ds | Files: %d/%d | Size: %s/%s",
		speedMBps,
		etaSec,
		sentFiles,
		totalFiles,
		transfer.FormatBytes(currentBytes),
		transfer.FormatBytes(totalBytes),
	)
}

func (b *broadcaster) renderFinalProgress(totalBytes int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// To avoid zero speed, we just print final state
	fmt.Printf("\r\033[KSpeed: %.1f MB/s | Pending: 0s | Files: Done | Size: %s/%s",
		b.lastSpeed,
		transfer.FormatBytes(totalBytes),
		transfer.FormatBytes(totalBytes),
	)
}
