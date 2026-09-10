# FileShare V2 Implementation Plan

Provide a brief description of the problem, any background context, and what the change accomplishes.
This plan addresses the implementation of V2 features requested for the FileShare application.

## User Review Required

- The Firewall persistent allow feature will require a UAC elevation prompt (Administrator privileges) on Windows to add the firewall rules. Is this acceptable to prompt the user during application startup, or would you prefer this to be a dedicated button in the Web UI/CLI command? I plan to add a CLI command `fileshare firewall` and a button in the Web UI to trigger it to avoid unexpected popups on startup.
- The default download directory will be hardcoded to `C:/Users/Public/Documents/FileShare`. 

## Proposed Changes

---

### store
Updates to the device storage to include MAC addresses.

#### [MODIFY] `pkg/store/store.go`
- Add `MACAddress string` to `LanInfo` struct.
- Update `RecordDevice` function to accept `macAddress` parameter and save it.

### discovery & server
Updates to device discovery and backend server to handle MAC address, new defaults, and download speed monitoring.

#### [MODIFY] `pkg/server/server.go`
- Change default download directory in `NewServer` to `C:/Users/Public/Documents/FileShare`.
- Update `handleInfo` to retrieve and include the server's local MAC address in the JSON response.
- Update `handleStreamUpload` and `handleMultipartUpload` to wrap `r.Body` with a custom tracking reader.
- Send periodic `transfer_progress` events over the SSE channel (`s.sseClients`) during download so the Web UI can show download speed.

#### [MODIFY] `pkg/discovery/peer.go`
- Update `ProbePeer` to parse the newly returned `macAddress` from `/api/info`.
- Pass the MAC address to `store.Default().RecordDevice()`.
- Extend `Peer` struct to include MAC Address.

### web UI
Updates to the frontend to support explicit URLs, copy to clipboard, and displaying download progress.

#### [MODIFY] `web/dist/index.html`
- Modify `#localDeviceIp` display to clearly show `http://...` and add a "Copy" icon button next to it.
- Add a "Configure Firewall" button in the settings/sidebar to trigger the UAC prompt for permanent firewall rules.

#### [MODIFY] `web/dist/js/app.js`
- Update the manual connection logic to strip `http://` or `https://` if a user pastes a full URL.
- Implement the "Copy URL" functionality for the local device URL.
- Add a listener in the WebSocket/SSE `ws.onmessage` to process `transfer_progress` events and update the `transferProgressCard` UI (which currently only works for uploads).

### cmd (CLI)
Updates to the CLI and main entry point.

#### [MODIFY] `cmd/fileshare/main.go`
- Add a new command `fileshare firewall` that executes an elevated PowerShell process to add the firewall rules: `New-NetFirewallRule ...`.
- Add an API endpoint in `server.go` to trigger this from the Web UI.

## Verification Plan

### Automated Tests
- Run `go build` to ensure all backend code compiles.

### Manual Verification
- Start the server using `fileshare ui` and check if the default directory is `C:/Users/Public/Documents/FileShare`.
- Check if the LAN URL is displayed in the UI and can be copied.
- Paste a full URL into the manual add field and verify it works.
- Send a file and verify upload speeds are shown.
- Send a file from a different device (or CLI) to the UI and verify download speeds are shown.
- Click the "Configure Firewall" button and verify a UAC prompt appears and rules are added.
- Check `store/devices.json` to verify MAC addresses are populated.
