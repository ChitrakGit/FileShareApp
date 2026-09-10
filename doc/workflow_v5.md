# FileShare V5 Workflow

## 1. Objective
Enhance the user experience of device connection during file transfers and expand the desktop application's native startup functionality to macOS and Linux.

## 2. Requirements Addressed
1. **On-the-fly Device Addition**: Provide a convenient button next to the "Send to Device" dropdown. Clicking this button allows the user to add a target device via its LAN IP address instantly without navigating to the sidebar.
2. **Cross-Platform Startup Support**: 
   - Extend the existing Windows registry startup feature to macOS (`~/Library/LaunchAgents/`) and Linux (`~/.config/autostart/`).
   - Implement `startup_darwin.go` and `startup_linux.go` to manage the native OS autostart mechanics.

## 3. Implementation Steps

### Step 1: Web UI Enhancement
- Modify `web/dist/index.html` to add a new `btnQuickAddDevice` next to the `targetPeerSelect` dropdown.
- Add a new modal `quickAddModal` to the HTML to capture the target IP and an optional name.
- Update `web/dist/js/app.js` to handle the modal opening, form submission via the existing `/api/peers/add` endpoint, and auto-select the newly added device in the dropdown upon success.

### Step 2: Cross-Platform Autostart Logic
- Create `pkg/desktop/startup_linux.go` with build tags for `linux`. It will generate a standard `.desktop` file pointing to the FileShare executable with the `desktop` command.
- Create `pkg/desktop/startup_darwin.go` with build tags for `darwin`. It will generate a standard `.plist` XML file for `launchd` to launch the FileShare executable on login.

### Step 3: Verification
- Verify that the application cross-compiles successfully for all three major operating systems.
- Ensure the quick-add feature accurately triggers the backend API and refreshes the UI dynamically.
