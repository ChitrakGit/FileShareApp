# FileShare V2 Implementation Plan

This document outlines the workflow and implementation plan for the requested FileShare V2 features.

## 1. Device Info & MAC Address Storage
- **Goal:** Save connected devices with their ID, LAN info, and MAC address into a JSON file in the `store` directory.
- **Implementation:**
  - `pkg/store/store.go` already saves a JSON registry (`store/devices.json`). We will extend the `LanInfo` struct to include a `MACAddress` field.
  - Update `pkg/server/server.go` (`handleInfo`) to look up and return the primary local MAC address alongside the IP/Port.
  - Update `pkg/discovery/peer.go` to parse the incoming MAC address and pass it to the store.
  - Update `RecordDevice` in `store.go` to accept and save the MAC address.

## 2. Direct Device Addition & URL Sharing
- **Goal:** Allow users to add a device via LAN URL with Port, and clearly display this URL so they can share it.
- **Implementation:**
  - **UI Changes (`web/dist/index.html` & `app.js`):** Modify the top Navigation Bar to display the full sharing URL (e.g., `http://192.168.1.15:8990`) instead of just the IP. Add a quick "Copy URL" button.
  - **Manual Connection:** Update `app.js`'s "Add Device" input parsing so that if a user pastes a full URL (like `http://192.168.1.5:8990`), it automatically strips the `http://` and extracts just the IP/Port to pass to the backend.

## 3. Real-Time Upload and Download Speeds
- **Goal:** Show upload and download speeds during transfers.
- **Implementation:**
  - **Upload:** The web UI (`app.js`) currently already calculates and displays the upload speed and ETA for the sender using `XMLHttpRequest.upload.onprogress`. We will ensure it functions smoothly.
  - **Download:** Currently, the receiver blocks until the file is fully received. We will wrap the incoming HTTP request body (`r.Body` in `handleStreamUpload` and `handleMultipartUpload`) with a custom `io.Reader` (like an `io.TeeReader` or a tracking reader) that periodically calculates the bytes read per second.
  - We will send these progress updates to the frontend via the existing Server-Sent Events (SSE) `/events` endpoint.
  - Update `app.js` to listen for the `transfer_progress` SSE messages and trigger the `transferProgressCard` UI for the receiving side to display live download speeds.

## 4. Default Save Directory
- **Goal:** Default download folder should be `C:/Users/Public/Documents/FileShare`.
- **Implementation:**
  - In `pkg/server/server.go` (`NewServer`), we will change the default fallback logic when no directory is provided to point to `C:\Users\Public\Documents\FileShare` instead of the user's `Downloads` folder.

## 5. Permanent Firewall Permissions
- **Goal:** Permanently allow the application through the firewall so the prompt doesn't repeatedly appear.
- **Implementation:**
  - We will add an initialization check or a specific CLI command (e.g., `fileshare firewall`) that adds inbound rules to the Windows Defender Firewall.
  - It will execute an elevated PowerShell command (prompting UAC once):
    ```powershell
    Start-Process powershell -Verb RunAs -ArgumentList "-Command `"New-NetFirewallRule -DisplayName 'FileShare TCP' -Direction Inbound -LocalPort 8990 -Protocol TCP -Action Allow; New-NetFirewallRule -DisplayName 'FileShare UDP' -Direction Inbound -LocalPort 53535 -Protocol UDP -Action Allow`""
    ```
  - This ensures the rule is permanently added at the OS level.
