package transfer

import (
	"context"
	"fmt"

	quic_tls "fileshare/pkg/quic"
	"github.com/quic-go/quic-go"
)

// StartQUICListener starts a QUIC server listening for file streams.
func StartQUICListener(port int, downloadDir string, pin string, onProgress func(percent float64, speed float64, eta int), onComplete func(fileCount int, totalBytes int64)) error {
	tlsConf := quic_tls.GenerateTLSConfig()
	listener, err := quic.ListenAddr(fmt.Sprintf(":%d", port), tlsConf, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				continue
			}

			go handleQUICConnection(conn, downloadDir, pin, onProgress, onComplete)
		}
	}()
	return nil
}

func handleQUICConnection(conn *quic.Conn, downloadDir string, pin string, onProgress func(percent float64, speed float64, eta int), onComplete func(fileCount int, totalBytes int64)) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return
	}
	defer stream.Close()
	defer conn.CloseWithError(0, "")

	// Handshake / PIN validation could be done by reading first bytes.
	// For simplicity, assuming the stream is raw Tar data for now.
	// In production, we would send a small header containing PIN and length.

	// Since we need to know the total length for progress, we must parse a small header.
	// We'll implement a simple header reading here if needed, or rely on TarExtractor.

	// Proceed with Tar extraction over QUIC
	fileCount, totalBytes, err := ExtractTar(stream, downloadDir, nil)
	if err != nil {
		fmt.Printf("[QUIC] Transfer error: %v\n", err)
		return
	}

	if onComplete != nil {
		onComplete(fileCount, totalBytes)
	}
	
	// Send ACK to client
	_, _ = stream.Write([]byte{1})
}

