# FileShare V3 Workflow

## 1. Objective
Implement FileShare V3 requirements:
1. **QR Code Connect:** Enable users to scan a QR code from another user's screen to instantly connect to their device.
2. **Add Device via URL:** Robustly support copying and pasting the LAN URL (e.g., `http://192.168.1.15:8990`) into the "Add Device" input.
3. **Cross-Platform Compatibility:** Ensure full compatibility for Windows, macOS, and Linux (e.g., handling OS-specific default download paths).

## 2. Implementation Steps

### Step 1: Cross-Platform Compatibility Fixes
- **`pkg/server/server.go`**: Update the hardcoded `C:/Users/Public/Documents/FileShare` default download directory to be OS-aware. For Windows, retain the public directory. For Linux and macOS, fall back to `~/Downloads/FileShare`.

### Step 2: Add Device via URL Refinement
- **`web/dist/js/app.js`**: Verify and expand the regex parsing for the "Save Device" manual connection input so that it can smoothly handle `http://`, `https://`, and trailing slashes if a user pastes a full URL.
- **`web/dist/index.html`**: Update UI placeholder text to clearly communicate that URLs are accepted.

### Step 3: QR Code Scanning Feature
- **Dependency**: Download a lightweight offline HTML5 QR scanner library (e.g., `html5-qrcode.min.js`) into `web/dist/js/` to maintain the "Zero Cloud" requirement.
- **UI Update (`index.html`)**: 
  - Add a "Scan QR" button next to the manual add device section.
  - Add a scanning modal with a `<div id="reader"></div>` for the camera view.
- **Frontend Logic (`app.js`)**: 
  - On click, request camera permissions and start the scanner.
  - On successful scan, parse the URL, stop the camera, populate the IP address field, and automatically trigger the "Save Device" API call.
- **QR Code Content**: The existing "Mobile Pair" QR code currently generates the URL `http://<ip>:<port>`. We will reuse this same QR code for device-to-device pairing.

### Step 4: Testing & Verification
- **Test File**: Generate an integration test suite `test/e2e_v3_test.go` to explicitly verify OS-specific path resolution and manual peer addition logic.
- **Manual Verification**: Launch the Web UI, trigger the QR scanner, and verify camera stream parsing.

### Step 5: Documentation Update
- Update `README.md` to reflect V3 features upon successful test completion, in compliance with project rules.
