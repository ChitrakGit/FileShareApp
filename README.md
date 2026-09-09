# FileShare - High-Speed LAN & Wi-Fi File Sharing

A lightweight, high-performance file sharing application built in Go. Send individual files, batches of files, or complete nested directories across your local Wi-Fi or LAN with zero cloud dependencies.

Features both an intuitive **Command-Line Interface (CLI)** and a sleek, modern **Web UI** with drag-and-drop support, peer radar discovery, and QR code pairing for mobile devices.

---

## Features

- **🚀 Ultra-Fast LAN Speed**: Transfers files directly across your Wi-Fi or Ethernet at full local network bandwidth.
- **📁 Multi-File & Entire Folder Streaming**: Send entire folder structures recursively in one go. Data streams on the fly with zero temporary files created on disk.
- **📡 Automatic Peer Discovery**: Devices automatically discover each other over local UDP broadcasts without needing manual configuration.
- **📱 Instant Mobile Pairing**: Display a Wi-Fi QR code in the Web UI to let any smartphone on the network upload/download files via mobile browser without installing any app.
- **🛡️ Secure by Design**: Built-in path traversal guards (`pathGuard`), optional 6-digit session PINs, and explicit accept/decline permissions.
- **📦 Single Standalone Binary**: The entire Web UI (HTML, CSS, JavaScript) is embedded directly into the Go executable using `//go:embed`.

---

## Quick Start

### 1. Launch the Web UI
```bash
fileshare ui
```
This launches the local server, announces presence on Wi-Fi, and automatically opens `http://localhost:8990` in your browser.

Custom port and download directory:
```bash
fileshare ui --port 9000 --dir C:\MyReceivedFiles
```

### 2. Discover Devices on Wi-Fi (CLI)
```bash
fileshare scan
```
Scans your subnet and displays all active FileShare nodes:
```
Found 2 peer(s) on the network:
----------------------------------------------------------------------
NAME                     ADDRESS                OS        
----------------------------------------------------------------------
Alice-Laptop             192.168.1.42:8990      windows   
Home-Server              192.168.1.15:8990      linux     
----------------------------------------------------------------------
```

### 3. Send Files & Folders (CLI)
Send single or multiple files:
```bash
fileshare send report.pdf photo.jpg --to 192.168.1.42:8990
```

Send entire directories (preserves complete folder hierarchy):
```bash
fileshare send ./ProjectSource/ ./Data/ --to 192.168.1.42:8990
```

With security PIN:
```bash
fileshare send ./Confidential/ --to 192.168.1.42:8990 --pin 482910
```

### 4. Headless Receiver Daemon (CLI)
Run as a background listener on a server:
```bash
fileshare receive --dir ~/Downloads/FileShare
```

---

## Project Structure

```
FileShare/
├── cmd/
│   └── fileshare/
│       └── main.go           # CLI entry point (scan, send, receive, ui)
├── pkg/
│   ├── discovery/
│   │   ├── beacon.go         # UDP broadcast & listener
│   │   └── peer.go           # Peer registry & timeouts
│   ├── transfer/
│   │   ├── archiver.go       # Tar streaming for files & folders
│   │   ├── extractor.go      # Tar unpacker & path traversal security
│   │   └── progress.go       # Terminal progress bar & speed calculator
│   ├── server/
│   │   └── server.go         # HTTP API & Web UI static asset server
│   └── client/
│       └── client.go         # Streaming client with live progress bar
├── web/
│   ├── embed.go              # //go:embed embedding frontend into binary
│   └── dist/
│       ├── index.html        # Modern dashboard layout
│       ├── css/style.css     # Glassmorphism dark mode styling
│       └── js/app.js         # Dropzone, folder reader & WebSocket client
├── go.mod
└── README.md
```

---

## Building from Source

```bash
# Build standalone executable
go build -o fileshare.exe ./cmd/fileshare

# Cross-compile for Linux (Raspberry Pi, VPS)
GOOS=linux GOARCH=amd64 go build -o fileshare-linux ./cmd/fileshare

# Cross-compile for macOS
GOOS=darwin GOARCH=arm64 go build -o fileshare-macos ./cmd/fileshare
```
