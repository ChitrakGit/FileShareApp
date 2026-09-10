# FileShare V4 Workflow

## 1. Objective
Transform FileShare into a fully-fledged desktop application with native system integration, a desktop widget, and QUIC-based high-performance networking.

## 2. Requirements Addressed
1. **Desktop App & Icon**: Wrap the application as a desktop app with an icon from the `icon/` directory.
2. **System Tray & Startup**: Run in the taskbar (system tray) on close, and start on boot.
3. **Settings**: Checkbox to toggle "Run at startup" and "Minimize to taskbar".
4. **Taskbar Actions**: Right-click context menu in the system tray to close the app.
5. **Floating Drawer**: A frameless, always-on-top glassmorphism drawer in the top right corner. Hovering opens it to reveal connected devices (grouped by device type) for quick drag-and-drop file transfers with progress bars.
6. **QUIC Protocol**: Upgrade the underlying transfer protocol from HTTP/TCP to QUIC (HTTP/3) for improved performance.
7. **Local Network**: Retain peer-to-peer local network capabilities.
8. **Device Icons**: Differentiate devices by icon (PC, Tablet, Mobile) and color-code connection status (Green = Connected, Gray/Red = Disconnected).
9. **Device Grouping**: If multiple devices of the same type are connected, group them with a numeric badge.
10. **Installer**: Generate an installer so the app can be searched in the Start Menu.

## 3. Implementation Steps

### Step 1: QUIC Protocol Upgrade
- Integrate `github.com/quic-go/quic-go` into `pkg/server` and `pkg/client`.
- Generate ephemeral in-memory TLS certificates for local QUIC peering.
- Modify the drag-and-drop file transfer flow to use the QUIC stream instead of standard HTTP chunked requests.

### Step 2: Desktop Framework Integration (Wails)
- Initialize a **Wails** (`github.com/wailsapp/wails/v2`) configuration to wrap the existing Go backend and `web/dist` frontend.
- Utilize Wails to create two windows:
  1. **Main Dashboard**: The existing Web UI.
  2. **Floating Drawer**: A frameless, transparent, always-on-top window positioned at the top right of the screen.

### Step 3: System Tray and Autostart
- Integrate `github.com/getlantern/systray` for the taskbar icon and right-click context menu.
- Implement Windows Registry bindings (`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`) to handle the "Run at Startup" feature.

### Step 4: Floating Drawer UI (Glassmorphism & Drag-and-Drop)
- Create a new HTML/JS interface for the drawer window.
- Implement CSS backdrop-filter for the glassmorphism effect.
- Group discovered peers by OS/device type and display appropriate SVG icons (Desktop, Mobile, Tablet) with notification badges.
- Implement Wails native drag-and-drop dropzone on this window to trigger the QUIC transfer.

### Step 5: Installer Generation
- Setup an `icon/` directory with a standard `.ico` file.
- Use Wails' built-in NSIS installer script generation to create `FileShare-Setup.exe` that registers the app in the Windows Start Menu.

### Step 6: Testing & Verification
- Create `test/e2e_v4_test.go` to verify QUIC transport reliability.
- Manually test the system tray, startup registry, and floating drawer UX.
