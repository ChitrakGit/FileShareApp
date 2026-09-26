# Implementation Plan: V6 (Android) & V7 (PC) Features

This plan outlines the architecture to implement the new 3-page UI (Send, Received, Settings), checksum verification, and persistent settings. Because the FileShare application uses a unified Web UI for both PC (Wails) and Android (WebView), we can implement these features in the web frontend and Go backend, and **both platforms will automatically get them**. 

## User Review Required
> [!IMPORTANT]
> The Android app and PC app share the same HTML/CSS/JS frontend. Therefore, building these 3 pages (Send, Received, Settings) in the web UI will give you the requested Android bottom navigation and the PC side navigation simultaneously without needing to write Kotlin code for the UI. Please approve this unified approach.

## Open Questions
> [!WARNING]
> 1. **Device Type Detection:** Detecting if an Android device is a "Tablet" vs "Smart Phone" perfectly from Go or the WebView is notoriously difficult without using Java/Kotlin bindings. Is it acceptable to use CSS media queries and User-Agent sniffing in JavaScript to roughly estimate Mobile vs Tablet?
> 2. **Restarting Server:** Changing settings like the Multicast port requires restarting the underlying network listener. This will briefly disconnect the UI. Is it acceptable to show a "Reconnecting..." spinner in the UI while it restarts?
> 3. **Default Device Names:** Should default random device names use an adjective-noun format (e.g., `Fast-Cheetah`)?

---

## Proposed Changes

### 1. Unified Web Frontend Refactor (HTML/CSS/JS)

We will convert the single-page `index.html` into a 3-view Single Page Application (SPA).

#### [MODIFY] `web/dist/index.html`
- **Navigation Structure**: Add a responsive navigation bar. On mobile devices, it will snap to the bottom of the screen with 3 icons (Send, Received, Settings). On PCs, it will appear as a side or top navigation.
- **View - Send Page**: Keep the existing file drag & drop, the QR code (mobile pairing button), and the transfer progress with speeds.
- **View - Received Page**: A new view displaying all files/folders received by the user, parsed from a new `/api/history` backend endpoint.
- **View - Settings Page**: A new view implementing the requested accordion/sections:
  - **Storage**: Change Save Directory.
  - **Server**: Restart Server button, Update Multicast address & port.
  - **User**: Update Device Name, display Device Info (Model, OS, Type).
  - **Network**: Block ports for sending/receiving.
  - Right-aligned **Refresh buttons** and a **Set Default Settings** button.

#### [MODIFY] `web/dist/js/app.js` & `web/dist/css/style.css`
- Add JavaScript logic to switch between the 3 views.
- Add API calls to fetch and save settings, trigger restarts, and fetch receive history.
- Add CSS for the Bottom Navigation bar on mobile (`@media (max-width: 768px)`).

---

### 2. Go Backend: Checksums

#### [MODIFY] `pkg/transfer/tar.go` & `pkg/transfer/transfer.go`
- Integrate `crypto/sha256`. 
- **Send**: As files are read and written to the TAR stream, compute the SHA-256 hash using an `io.TeeReader`. 
- **Receive**: Compute the hash as the stream is extracted.
- After the file transfer completes, the sender will send a final verification manifest, and the receiver will compare the local hash against the sender's hash to ensure data integrity.

---

### 3. Go Backend: Settings & History APIs

#### [NEW] `pkg/store/settings.go` (or `pkg/settings`)
- Create a persistent JSON file (e.g., `config.json`) in the user's config directory to store: Save Directory, Device Name, Multicast Port/Address, and Blocked Ports.

#### [MODIFY] `pkg/server/server.go`
- Add `GET /api/settings` and `POST /api/settings`.
- Add `POST /api/settings/restart` to gracefully shutdown the listener and start it again with new ports.
- Add `GET /api/history` to scan the Save Directory and return a list of received files with metadata.

## Verification Plan
1. **Automated Tests**: I will run `go test ./test/...` to ensure that standard transfers still succeed and that the new checksum logic doesn't break directory streaming.
2. **Manual UI Verification**: We will open the `web/dist/index.html` locally in a mobile view to verify the bottom navigation bar and the settings accordion.
