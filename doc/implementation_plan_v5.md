# FileShare V5 Implementation Plan

This document outlines the implementation plan for the V5 features:
1. Adding a device on the fly while selecting a target peer.
2. Extending Desktop Startup support to Linux and macOS.

## User Review Required
> [!IMPORTANT]
> - Adding a device on the fly will introduce a small popup modal next to the "Send to Device" dropdown, allowing you to quickly type the IP address of a device on the same LAN. 
> - The macOS startup logic will use standard `~/Library/LaunchAgents` plist entries.
> - The Linux startup logic will use standard `~/.config/autostart` desktop entries.
> 
> Please review and click "Proceed" to approve this plan.

## Proposed Changes

### Web Application UI

#### [MODIFY] [index.html](file:///c:/Company%20Files/Study/FileShare/web/dist/index.html)
- Add a new "quick add" button `btnQuickAddDevice` next to the `targetPeerSelect` dropdown in the `send-form` area.
- Add a new modal dialog `quickAddModal` specifically tailored for quickly adding a device IP and Name on the fly, streamlining the user experience over scrolling to the right sidebar.

#### [MODIFY] [app.js](file:///c:/Company%20Files/Study/FileShare/web/dist/js/app.js)
- Bind an event listener to `btnQuickAddDevice` to display the new modal.
- Bind the submit action of the `quickAddModal` to call the existing `/api/peers/add` backend endpoint.
- Upon success, automatically select the newly added device in the `targetPeerSelect` dropdown.

### Desktop Startup (Cross-Platform)

#### [NEW] [startup_linux.go](file:///c:/Company%20Files/Study/FileShare/pkg/desktop/startup_linux.go)
- Implement `EnableStartup()`, `DisableStartup()`, and `IsStartupEnabled()` using `~/.config/autostart/fileshare.desktop`. 
- `EnableStartup()` will create the directory if it doesn't exist, and write a standard `.desktop` file executing `fileshare desktop`.

#### [NEW] [startup_darwin.go](file:///c:/Company%20Files/Study/FileShare/pkg/desktop/startup_darwin.go)
- Implement the startup functions using macOS `~/Library/LaunchAgents/com.fileshare.app.plist`.
- `EnableStartup()` will write the property list XML configuring the `ProgramArguments` to run `fileshare desktop` and `RunAtLoad` to true.

## Verification Plan

### Automated Tests
- The E2E tests for V3 and V4 will continue to run to ensure there are no regressions.

### Manual Verification
- Launch the UI locally and verify that clicking the new Add Device button opens the modal and correctly saves a mock device, automatically selecting it in the dropdown.
- Check the syntax and paths of the `startup_linux.go` and `startup_darwin.go` implementations to guarantee cross-platform compilation works seamlessly alongside `startup_windows.go`.
