# FileShare Implementation Walkthrough

A lightweight, high-performance, cross-platform file sharing application built in Go for local networks (LAN & Wi-Fi).

---

## 1. Summary of Accomplishments

1. **Installed Go (Golang 1.27.0)** via `winget` on the system.
2. **Built the Core Streaming Transfer Engine** (`pkg/transfer`):
   - On-the-fly streaming tar archiver (`archiver.go`) capable of packaging individual files, batches of files, and recursive directory trees with zero disk overhead.
   - Streaming tar extractor (`extractor.go`) with rigorous **Path Traversal Guards** preventing malicious directory navigation (`../`).
   - Progress reader/writer (`progress.go`) providing live terminal progress bars with percentage completion, transfer rate (MB/s), and estimated time remaining (ETA).
3. **Implemented Automatic LAN/Wi-Fi Discovery** (`pkg/discovery`):
   - UDP broadcast beaconing (`beacon.go`) announcing node availability, IP, and hostname every 2 seconds.
   - Dynamic peer registry (`peer.go`) tracking active devices and automatically pruning stale nodes.
4. **Built HTTP Transfer Server & Embedded Web UI** (`pkg/server` & `web`):
   - High-throughput streaming upload endpoints (`/api/stream-upload` for CLI and `/api/upload` for browser multi-file/folder drag & drop).
   - Entire frontend (HTML, CSS, JS) embedded directly into the single binary via `//go:embed`.
   - Peer radar listing, drag-and-drop dropzone, live upload progress meters, and mobile QR code pairing.
5. **Created the CLI Suite** (`cmd/fileshare`):
   - `fileshare ui`: Starts server, announces presence, and opens browser.
   - `fileshare scan`: Scans LAN for 3 seconds and displays active devices in a formatted table.
   - `fileshare send <paths...> --to <target>`: Sends files and directories with live terminal progress.
   - `fileshare receive --dir <path>`: Runs a headless receiver daemon.

---

## 2. Test & Verification Results

### Unit Tests
```bash
go test -v ./...
```
```
=== RUN   TestPeerRegistry
--- PASS: TestPeerRegistry (0.10s)
PASS
ok      fileshare/pkg/discovery 0.430s
=== RUN   TestStreamAndExtractTarRoundTrip
--- PASS: TestStreamAndExtractTarRoundTrip (0.02s)
=== RUN   TestPathTraversalSecurity
--- PASS: TestPathTraversalSecurity (0.00s)
PASS
ok      fileshare/pkg/transfer  0.390s
```

### End-to-End Transfer Test
A nested directory (`test_source_dir`) containing a root file and a deep nested subfolder (`nested_folder\inner.txt`) was sent across local ports using the compiled binary:
```bash
.\fileshare.exe send .\test_source_dir --to 127.0.0.1:8995
```
Output:
```
[Sender] Preparing to send 4 items (49 B) to 127.0.0.1:8995...
[██████████████████████████████] 100.0% | 49 B/49 B | 0.0 MB/s | ETA: 0s
[Sender] Transfer completed successfully in 7ms!

Peer saved 2 items.
```
Verified that all files, directory structures, and file contents were extracted byte-for-byte identically.

---

## 3. How to Use

### Run Web UI
```bash
.\fileshare.exe ui
```
Opens your browser at `http://localhost:8990` and announces your device on Wi-Fi.

### Scan Peers on Local Network
```bash
.\fileshare.exe scan
```

### Send Files or Entire Folders via CLI
```bash
.\fileshare.exe send ./MyFolder/ document.pdf --to 192.168.1.50:8990
```

### Receive Files in Daemon Mode
```bash
.\fileshare.exe receive --dir C:\Users\<YourUser>\Downloads\FileShare
```
