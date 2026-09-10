# 🚀 FileShare

> High-speed, peer-to-peer file and folder sharing over local Wi-Fi and LAN networks. Built with Go for maximum throughput and zero cloud dependencies.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux%20%7C%20Mobile-blue)](https://github.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📖 Table of Contents
- [Features](#-features)
- [Architecture & How It Works](#-architecture--how-it-works)
- [How to Build from Source](#-how-to-build-from-source)
- [Quick Start Guide](#-quick-start-guide)
  - [1. Web UI Mode (Recommended)](#1-web-ui-mode-recommended)
  - [2. CLI Mode](#2-cli-mode)
- [Working Across Two Routers (Subnets)](#-working-across-two-routers-subnets)
- [CLI Command Reference](#-cli-command-reference)
- [Security & Safety](#-security--safety)
- [Troubleshooting & FAQs](#-troubleshooting--faqs)

---

## ✨ Features

- **⚡ Full Wire & Wi-Fi Speed**: Transfers files directly between devices on your local network at maximum bandwidth (up to 1 Gbps+ on Ethernet/Wi-Fi 6) without routing through any external cloud server.
- **📁 Multi-File & Entire Directory Streaming**: Send single files, batches of selected files, or complete nested folder hierarchies in one go. Data streams on the fly with **zero temporary files created on disk**.
- **💻 Dual Interface (CLI & Web UI)**:
  - **Modern Web Dashboard**: Glassmorphic dark-mode UI with drag-and-drop dropzone, live real-time progress gauges (Upload & Download MB/s, ETA, % completion), and nearby device radar.
  - **Power-User CLI**: Scriptable terminal commands with animated progress bars for servers and headless machines.
- **📱 Instant Pairing (QR Code & URL)**:
  - **Mobile Access**: Scan the Wi-Fi QR code from your phone's camera to open FileShare on your mobile browser (iOS/Android) and send/receive files without installing any app.
  - **Device-to-Device QR Pairing**: Click "Scan QR Code" in the Web UI to instantly scan another screen and connect two devices without typing IP addresses!
  - **Simple URL Connect**: Easily copy your LAN URL (e.g. `http://192.168.1.100:8990`) and paste it straight into the Add Device input to instantly connect.
- **📡 Smart Auto-Discovery & Persistent Device Manager**:
  - Automatically discovers peers on the same subnet using UDP beacons (`53535`).
  - Active subnet scanner (`fileshare scan --subnet 192.168.1`) for multi-router setups.
  - Permanent device registry (`fileshare add`) so you can send files by **device name** (e.g. `--to "Office-PC"`) without remembering IPs. Tracks MAC Addresses for reliable device identification.
- **📦 Single Standalone Executable (Cross-Platform)**:
  - Compiled into a single binary (`fileshare.exe` on Windows, or native executables on macOS and Linux) with the entire Web UI baked inside via `//go:embed`. No Node.js, Python, or runtime dependencies required. Automatically resolves correct native OS download paths.
- **🛡️ Built-in Security**:
  - Strict path traversal guards (`pathGuard`) to prevent malicious `../` overwrites.
  - Optional 6-digit session PIN authentication.

---

## 🆕 Version History

### What's New in V5
- **Cross-Platform Autostart**: The desktop application now natively supports "run at startup" on Windows, macOS, and Linux.
- **Quick Device Pairing**: You can instantly add a device on your local network using the new "Add Device" button right next to the target selector.

### What's New in V4 (Desktop App Edition)
- **Native Desktop App**: Packaged with Wails for a full native desktop experience.
- **System Tray Integration**: Run FileShare silently in the background from your taskbar/system tray.
- **Run at Startup**: Option to automatically start the app on system boot.
- **Floating Drawer UI**: A sleek, glassmorphic floating drawer that appears on hover when in background mode.
- **QUIC Protocol**: Upgraded transport layer to use QUIC for even faster, multiplexed, and encrypted transfers.
- **Smart Device Grouping**: Automatically identifies and groups devices by type (PC, Tablet, Mobile) with appropriate icons in the UI.

### What's New in V3 (Ease of Use)
- **QR Code Pairing**: Share and scan QR codes to instantly pair devices.
- **URL Connection**: Add devices simply by copy-pasting their FileShare LAN URL.
- **Cross-Platform Compatibility**: Connect seamlessly between Linux, macOS, and Windows.

---

## 🏗 Architecture & How It Works

```
 ┌────────────────────────────────────────────────────────┐
 │                      User Layer                        │
 │    CLI (`fileshare`)    │    Web UI / Mobile Browser   │
 └───────────┬─────────────┴────────────────┬─────────────┘
             │                              │
 ┌───────────┴──────────────────────────────┴─────────────┐
 │                   Application Core                     │
 │  ┌────────────────────┐ ┌───────────────────────────┐  │
 │  │  Discovery Engine  │ │      Transfer Engine      │  │
 │  │ - UDP Beacon (LAN) │ │ - On-the-fly Tar Stream   │  │
 │  │ - Active Subnet    │ │ - Chunked HTTP Streaming  │  │
 │  │ - Persistent Store │ │ - Live Progress & ETA     │  │
 │  └────────────────────┘ └───────────────────────────┘  │
 │  ┌────────────────────┐ ┌───────────────────────────┐  │
 │  │   Security Layer   │ │      HTTP / WS Server     │  │
 │  │ - Path Guard       │ │ - Embedded Static Assets  │  │
 │  │ - 6-Digit PIN Auth │ │ - REST API & Event Stream │  │
 │  └────────────────────┘ └───────────────────────────┘  │
 └───────────────────────────┬────────────────────────────┘
                             │ Local Wi-Fi / LAN Network
 ┌───────────────────────────┴────────────────────────────┐
 │                      Remote Peer                       │
 └────────────────────────────────────────────────────────┘
```

---

## 🛠 How to Build from Source

### Prerequisites
- [Go 1.22 or newer](https://go.dev/dl/) installed. Verify with:
  ```bash
  go version
  ```

### 1. Build for Windows
```powershell
# Open terminal in project root
cd "C:\Company Files\Study\FileShare"

# Build standalone executable
go build -o fileshare.exe .
```

### 2. Cross-Compile for Other Platforms
You can compile binaries for other devices directly from your machine:

```powershell
# Linux (Ubuntu, Debian, Raspberry Pi, Home Server)
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o fileshare-linux .

# macOS (Apple Silicon M1/M2/M3)
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o fileshare-macos-arm64 .

# macOS (Intel)
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o fileshare-macos-intel .
```

---

## 🚀 Quick Start Guide

### 1. Web UI Mode (Recommended)

Simply run:
```powershell
.\fileshare.exe ui
```
1. FileShare starts the local server and **automatically opens your browser** at `http://localhost:8990`.
2. Other computers on the same network can access it at `http://<your-ip>:8990`. You can easily copy your full LAN URL directly from the top navigation bar.
3. **To connect a smartphone**: Click the **Mobile Pair** button in the top right and scan the QR code with your phone's camera.
4. **Drag & Drop**: Drop individual files or entire folders into the dropzone, choose a target device, and click **Send**.

Custom port and custom download directory (Default is `C:\Users\Public\Documents\FileShare`):
```powershell
.\fileshare.exe ui --port 9000 --dir "C:\MyDownloads"
```

---

### 2. CLI Mode

#### A. Discover Devices on the Network
```powershell
.\fileshare.exe scan
```
Output:
```
Found 2 peer(s) on the network:
--------------------------------------------------------------------------------
NAME                     ADDRESS                OS         SAVED   
--------------------------------------------------------------------------------
Alice-MacBook            192.168.1.42:8990      darwin     No      
Home-Server              192.168.1.15:8990      linux      Yes     
--------------------------------------------------------------------------------
```

#### B. Send Files or Full Folders
```powershell
# Send single or multiple files
.\fileshare.exe send ./report.pdf ./photo.jpg --to 192.168.1.42:8990

# Send an entire directory (preserves complete nested folder structure)
.\fileshare.exe send ./ProjectSource/ ./Datasets/ --to 192.168.1.42:8990
```

Live terminal progress bar:
```
[Sender] Preparing to send 48 items (320.5 MB) to 192.168.1.42:8990...
[████████████████████████░░░░░░]  78.4% | 251.2 MB/320.5 MB | 46.8 MB/s | ETA: 2s
```

#### C. Run as a Headless Receiver (Server Daemon)
```powershell
.\fileshare.exe receive --dir "C:\Downloads\FileShare"
```

---

## 🌐 Working Across Two Routers (Subnets)

If your network has two routers (e.g. **Router A** is connected to the Internet, and **Router B** is connected to Router A):

```
       [ Internet ]
            │
      [ Router A ] ─── Device A (IP: 192.168.1.50)
            │
      [ Router B ] ─── Device B (IP: 192.168.2.75)
```

Because consumer routers block UDP broadcasts between subnets, follow these simple steps:

### Option 1: Save Device Once (Easiest)
Save the target machine permanently so you never have to type its IP again:
```powershell
# On Device B: Save Device A
.\fileshare.exe add 192.168.1.50:8990 Device-A

# Send files anytime directly by name:
.\fileshare.exe send ./my-files/ --to Device-A
```

### Option 2: Active Subnet Scan
Scan the adjacent router's subnet in 1 second using active probing:
```powershell
# From Device B, scan Router A's subnet:
.\fileshare.exe scan --subnet 192.168.1

# From Device A, scan Router B's subnet:
.\fileshare.exe scan --subnet 192.168.2
```

### Option 3: Access Point Mode (Permanent Fix)
In Router B's settings, switch operation mode from **Router Mode** to **Access Point (AP) Mode**. Both routers will now share the same subnet (`192.168.1.x`), and automatic discovery will work seamlessly.

---

## 📋 CLI Command Reference

| Command | Description | Example |
| :--- | :--- | :--- |
| `fileshare ui` | Launch server & open browser Web UI | `fileshare ui --port 8990 --dir ./recv` |
| `fileshare scan` | Scan local Wi-Fi/LAN for active peers | `fileshare scan` |
| `fileshare scan --subnet <prefix>` | Actively probe all 254 IPs on a subnet | `fileshare scan --subnet 192.168.1` |
| `fileshare add <IP:port> [name]` | Verify and save remote device permanently | `fileshare add 192.168.1.50:8990 Laptop-A` |
| `fileshare list` | List all saved and discovered peers | `fileshare list` |
| `fileshare remove <name\|IP>` | Remove a device from saved registry | `fileshare remove Laptop-A` |
| `fileshare send <paths...> --to <dest>` | Send files/folders to IP or device name | `fileshare send ./photos/ --to Laptop-A` |
| `fileshare receive` | Run headless receiver daemon | `fileshare receive --dir ~/Downloads` |
| `fileshare firewall` | Permanently allow FileShare in Windows Firewall | `fileshare firewall` |

---

## 🔒 Security & Safety

1. **Path Traversal Protection**:
   All incoming files are parsed using `filepath.Clean`. FileShare actively blocks any path containing `../` or root paths (`/` or `C:\`), guaranteeing files can only be saved inside the target download directory.
2. **Security PIN (Optional)**:
   Protect transfers on public or office Wi-Fi by requiring a 6-digit PIN:
   ```powershell
   # Receiver requires PIN 482910
   .\fileshare.exe receive --pin 482910

   # Sender must provide matching PIN
   .\fileshare.exe send ./secret.zip --to 192.168.1.42:8990 --pin 482910
   ```

---

## 🔧 Troubleshooting & FAQs

### 1. "Device not found" during `fileshare scan`
- **Cause**: Router AP Isolation or devices connected across different router subnets.
- **Solution**:
  - Run active subnet scan: `fileshare scan --subnet 192.168.1`
  - Or add the device directly: `fileshare add <device-ip>:8990 MyDevice`

### 2. Windows Defender Firewall blocks connections
- **Cause**: Windows may prompt to allow network access when running for the first time.
- **Solution**: 
  - **Easiest**: Click the **"Configure Firewall (Windows)"** button in the Web UI or run `fileshare firewall` in the CLI to automatically add persistent rules (requires Administrator privileges).
  - **Manual**: Click **"Allow access"** on private networks when prompted. Alternatively, add a firewall rule via Administrator PowerShell:
  ```powershell
  New-NetFirewallRule -DisplayName "FileShare Port 8990" -Direction Inbound -LocalPort 8990 -Protocol TCP -Action Allow
  New-NetFirewallRule -DisplayName "FileShare Discovery 53535" -Direction Inbound -LocalPort 53535 -Protocol UDP -Action Allow
  ```

### 3. Port 8990 is already in use
- **Solution**: Specify a different port:
  ```powershell
  .\fileshare.exe ui --port 9090
  ```

### 4. Mobile phone cannot open Web UI
- **Check 1**: Make sure your phone is connected to the **same Wi-Fi network** as your computer (not mobile cellular data).
- **Check 2**: Ensure you typed the computer's local Wi-Fi IP (e.g. `http://192.168.1.15:8990`), **not** `localhost`. Scanning the QR code automatically uses the correct IP.

### 5. Transfers of very large folders (10 GB+)
- FileShare uses HTTP streaming backpressure with `io.Pipe` and chunked transfer encoding. Even multi-gigabyte files will transfer smoothly without filling your computer's RAM. Ensure the destination drive has sufficient free disk space.

---

## 📄 License
This project is open-source software licensed under the **MIT License**.
