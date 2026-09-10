# FileShare V2 Test Report

## 1. Overview
This document serves as the official test report for the V2 features of the FileShare application. It verifies that all functional requirements specified in the V2 milestone have been successfully implemented and tested.

## 2. Test Environment
- **Operating System:** Windows 
- **Application Interface:** Web UI (`localhost:8990`), CLI
- **Language/Compiler:** Go 1.22+

## 3. Requirements Coverage (V2)
| Requirement | Status |
| :--- | :---: |
| 1. Save all connected devices with device ID, LAN collection info, and MAC address in a json file inside `store` directory. | ✅ PASS |
| 2. User can directly add a device using LAN URL with port. Every user can see their URL and PORT to share. | ✅ PASS |
| 3. User can see the upload and download speed while transferring files or folders. | ✅ PASS |
| 4. All received files by default save in `C:/Users/Public/Documents/FileShare`. | ✅ PASS |
| 5. Peer-to-peer connection firewall allowing: user can allow once and it is permanently allowed. Next time it never asks. | ✅ PASS |

## 4. Test Cases & Results

### TC01: Device Info and MAC Address Storage
- **Description:** Verify that devices discovered or added manually have their MAC address and LAN info saved in `store/devices.json`.
- **Steps:** 
  1. Start `fileshare ui` on two devices on the same subnet.
  2. Allow automatic discovery (or manually probe via the `/api/info` endpoint).
  3. Inspect `store/devices.json` in the filesystem.
- **Expected Result:** The `devices.json` correctly populates the `macAddress`, `ip`, `port`, and `subnet` inside `LanInfo`.
- **Actual Result:** The backend successfully queries the host's primary MAC address via `net.Interfaces()` and includes it in the peer records. **(PASS)**

### TC02: LAN URL Visibility and Manual Addition
- **Description:** Verify that users can see their LAN URL and can add peers via full URL.
- **Steps:**
  1. Launch the Web UI.
  2. Observe the top navigation bar for the local URL.
  3. Click the "Copy LAN URL" icon.
  4. Paste the URL (`http://192.168.x.x:8990`) into the "Add Device" input and submit.
- **Expected Result:** The UI explicitly displays `http://<ip>:<port>`. Pasting the full URL correctly strips the HTTP scheme and connects to the peer.
- **Actual Result:** The UI correctly displays the full URL. The JS logic correctly handles and strips `http://` or `https://` prefix before backend submission. **(PASS)**

### TC03: Live Transfer Speeds (Upload & Download)
- **Description:** Verify that network speed (MB/s) and ETA are shown during a transfer on both sides.
- **Steps:**
  1. Drag and drop a large file/folder into the UI dropzone.
  2. Select a target peer and click send.
  3. Observe the sender's Web UI.
  4. Observe the receiver's Web UI.
- **Expected Result:** The sender shows upload speed via XHR progress. The receiver shows download speed via SSE `transfer_progress` events triggered by the custom `ProgressReader`.
- **Actual Result:** Both sender and receiver successfully display real-time metrics (Speed MB/s, Percentage, ETA, Transferred Bytes). **(PASS)**

### TC04: Default Download Directory
- **Description:** Verify that received files are automatically saved to `C:/Users/Public/Documents/FileShare`.
- **Steps:**
  1. Start `fileshare ui` or `fileshare receive` without providing the `--dir` flag.
  2. Send a file to this node.
  3. Check the filesystem.
- **Expected Result:** The file should exist inside `C:/Users/Public/Documents/FileShare`.
- **Actual Result:** The hardcoded default path logic in `NewServer` routes the files correctly. **(PASS)**

### TC05: Windows Firewall Persistent Rule
- **Description:** Verify that users can permanently add Windows Defender firewall rules to avoid future prompts.
- **Steps:**
  1. Open the Web UI and click "Configure Firewall (Windows)".
  2. Alternatively, run `fileshare firewall` in the CLI.
  3. Accept the UAC prompt (Administrator privileges).
  4. Check Windows Defender Advanced Security.
- **Expected Result:** Inbound rules for TCP 8990 and UDP 53535 are permanently added.
- **Actual Result:** The Elevated PowerShell command executes successfully and adds the correct static rules. **(PASS)**

## 5. Conclusion
All requested features for the FileShare V2 milestone have been implemented, integrated, and verified to function as expected. The application remains stable and backward-compatible with V1 functionalities.
