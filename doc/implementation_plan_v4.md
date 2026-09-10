# FileShare V4 Implementation Plan

The V4 update is a massive architectural evolution. It transitions FileShare from a CLI/Local-Server application to a native, installable Desktop Application with advanced networking.

## Proposed Architecture & Frameworks

To satisfy the requirements of a floating, glassmorphism desktop drawer and system tray integration while reusing our beautiful Web UI, we will adopt **Wails v2** (`github.com/wailsapp/wails/v2`). Wails allows us to build native desktop applications using Go for the backend and our existing HTML/JS/CSS for the frontend.

For the networking layer, we will integrate **quic-go** (`github.com/quic-go/quic-go`) to upgrade the transfer protocol to QUIC (HTTP/3 over UDP).

---
## User Review Required

> [!WARNING]
> **Major Architectural Shift**
> Adopting Wails means the build process will change. We will no longer just run `go build`. Instead, we will use the Wails CLI (`wails build`) which packages the app into a native `FileShare.exe` complete with an installer, an icon, and WebView2 embedding. Is this acceptable?

> [!IMPORTANT]
> **QUIC Protocol and TLS**
> QUIC strictly requires TLS encryption. To keep the app zero-configuration and local, we will automatically generate in-memory, self-signed TLS certificates when the app starts. The clients will be configured to trust these local certificates automatically. 

---
## Proposed Changes

### 1. Desktop & UI Integration
- **Framework**: Wrap the application using Wails. The main Web UI will become the primary application window.
- **System Tray**: Integrate `github.com/getlantern/systray`. When the user closes the main window, it will hide to the tray instead of exiting. Right-clicking the tray icon will offer an "Exit" option.
- **Settings / Autostart**: Add a settings menu in the Web UI to toggle "Run at startup" and "Minimize to tray". The backend will manipulate the Windows Registry (`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`) to enforce startup.

### 2. Floating Glassmorphism Drawer
- We will spawn a *second*, frameless, transparent Wails window positioned at the top-right of the user's screen.
- **UI Design**: It will use CSS `backdrop-filter: blur(10px)` for the glassmorphism effect. It will stay collapsed and expand on mouse hover.
- **Device Grouping**: The drawer will display discovered peers. It will group them by OS (PC, Mobile, Tablet) and show numeric badges (e.g., a Mobile icon with a green "2" badge).
- **Drag & Drop**: Users can drag files directly onto these device icons in the drawer to instantly start a transfer.

### 3. QUIC Networking Integration
- Add `github.com/quic-go/quic-go` to the project.
- Implement a new `pkg/transfer/quic_server.go` and `quic_client.go`.
- File chunks will be streamed over QUIC multiplexed streams instead of standard TCP HTTP requests. This eliminates TCP head-of-line blocking and maximizes LAN throughput.

### 4. Installer and Icons
- We will place a high-quality icon in the `icon/` directory.
- We will configure `wails.json` to generate an NSIS installer (`FileShare-Setup.exe`). Installing this will register "FileShare" in the Windows Start Menu search.

## Verification Plan

### Automated Tests
- Create `test/e2e_v4_test.go` to specifically mock and verify the QUIC TLS handshake and stream transfer between two local instances.

### Manual Verification
- Compile the app using Wails.
- Install the app using the generated installer.
- Verify Start Menu searchability.
- Verify the app launches on Windows boot.
- Verify the floating drawer appears, expands on hover, and successfully accepts dropped files.
