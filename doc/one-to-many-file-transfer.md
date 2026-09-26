# One-To-Many File Transfer Architecture

## Overview
This document outlines the architecture for one-to-many file transfers in FileShare. The goal is to allow a single sender to dispatch files to multiple receivers concurrently, while maintaining reliability and efficiency.

## Functional Requirements
- **1-to-N Dispatch:** A single sender can broadcast files to multiple target addresses.
- **Resilience to Slow Receivers:** If one receiver is significantly slower than others, or the connection is poor (e.g. takes longer than 10 seconds to write a chunk), that specific receiver is dropped. The transfer to the remaining healthy receivers continues uninterrupted.
- **Progress Tracking:** The CLI interface shows real-time transfer progress displaying speed (MB/s), pending time, files sent / total files, and data sent / total data.
- **Data Integrity:** File checksums are transmitted to ensure the receiving devices can verify the downloaded files.

## Connection Lifecycle
1. **Discovery:** Receivers broadcast their presence via UDP on the LAN. The Sender uses the CLI or UI to discover available receivers.
2. **Handshake (QUIC):** The Sender iterates over the target receiver addresses and establishes QUIC connections concurrently.
3. **Data Transfer:** 
   - Files are packed into a tar stream on the fly.
   - The Sender uses a custom broadcaster to pipe this stream to all active QUIC connections.
   - Flow control is handled by QUIC. If a receiver's receive buffer fills up and write operations block for more than 10 seconds, the sender explicitly drops that connection.
4. **Completion:** Upon finishing the stream, the sender closes each connection and awaits an acknowledgment. Dropped receivers are not awaited.

## Data Transfer Protocol
- **Chunking and Streaming:** Files are streamed as tar archives chunk-by-chunk without loading the entire payload into memory.
- **Bounded Buffers / Timeouts:** To prevent memory bloat, `io.Pipe` is used alongside time-bounded goroutine writes. If a single write to a specific stream exceeds the specified timeout (`10s`), the connection is severed.
- **End of Stream Checksums:** As the final piece of data in the tar archive, a special JSON file `.fileshare_checksums.json` is appended containing SHA-256 hashes for all sent files. The receiver can use this to verify integrity.

## Security & Error Handling
- **Partial Failure:** Handled seamlessly through the custom multi-broadcaster (`broadcaster` inside `pkg/client`). When device B receives but device C drops, device B is unaffected.
- **Transport Security:** All data traverses the network using QUIC streams encrypted with ephemeral TLS certificates, preventing eavesdropping on LANs.
- **Path Traversal Protection:** The receiver unpacks the incoming tar stream safely, checking all relative paths and explicitly preventing directory traversal attacks.
