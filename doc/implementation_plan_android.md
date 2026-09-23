# Android Application Implementation Plan

This document outlines the detailed architecture and step-by-step plan for building the FileShare Android app. By leveraging your existing Go backend and embedded Web UI, we can build a highly maintainable, fully-featured native Android app with minimal Kotlin code.

## User Review Required
> [!IMPORTANT]
> Please review this architecture. The primary strategy uses **`gomobile`** to compile the Go backend into a native Android library (`.aar`). The Android app will start this Go server in the background and display your existing glassmorphic Web UI inside a full-screen `WebView`. This ensures 100% feature parity (QUIC transfers, peer discovery, UI updates) without needing to rewrite logic in Kotlin. 

## Open Questions
> [!WARNING]
> 1. Do you prefer a **fully native Kotlin UI** (Jetpack Compose) over the embedded WebView approach? (Note: A native UI would require rewriting all UI elements and communicating with the Go backend via API/Bindings, significantly increasing development time).
> 2. Have you already installed the Android SDK and `gomobile` toolchain on your machine?

---

## Proposed Architecture

```mermaid
graph TD
    A[Android App (Kotlin)] -->|onCreate() starts| B[Go Mobile Binding]
    B -->|Starts| C[Go HTTP & QUIC Server]
    A -->|Loads| D[WebView UI]
    D -->|Connects to| C[localhost:8990]
    C -->|Auto-Discovery| E[Local Network]
    C -->|QUIC Transfers| F[Remote Peers]
```

## Proposed Changes

### 1. Go Mobile Wrapper

#### [NEW] `pkg/mobile/mobile.go`
We will create a small wrapper package designed explicitly for `gomobile`. This package will expose simple `StartServer` and `StopServer` functions that Android's Java/Kotlin code can call directly.

```go
package mobile

import (
	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

var srv *server.Server
var disc *discovery.Service

// StartServer initializes the background Go server for the Android app
func StartServer(port int, downloadDir string) error {
	reg := discovery.NewPeerRegistry()
	var err error
	disc, err = discovery.NewService(port, reg)
	if err != nil {
		return err
	}
	disc.Start()
	
	srv = server.NewServer(port, downloadDir, disc, true, "")
	go srv.Start()
	return nil
}

// StopServer cleanly tears down the Go server
func StopServer() {
	if srv != nil {
		srv.Stop()
	}
	if disc != nil {
		disc.Stop()
	}
}
```

### 2. Android Project Setup

#### [NEW] `android/` Directory
We will use the Android CLI to generate a standard Android project structure inside the `android/` folder.
- **Build tool**: Gradle
- **Language**: Kotlin

### 3. Generate Android Archive (.aar)

We will use the `gomobile bind` command to compile the `pkg/mobile` Go code into an Android library format and place it into the Android project's `libs/` folder.
```bash
gomobile bind -target=android -o android/app/libs/fileshare.aar ./pkg/mobile
```

### 4. Android App Implementation

#### [MODIFY] `android/app/src/main/AndroidManifest.xml`
- Add necessary network permissions (`INTERNET`, `ACCESS_WIFI_STATE`, `ACCESS_NETWORK_STATE`, `CHANGE_WIFI_MULTICAST_STATE`).
- Add storage permissions (`READ_EXTERNAL_STORAGE`, `WRITE_EXTERNAL_STORAGE`) to allow the Go server to save files.

#### [MODIFY] `android/app/src/main/java/.../MainActivity.kt`
- Override `onCreate` to call `Mobile.startServer(8990, getExternalFilesDir(null).getAbsolutePath())`.
- Initialize a full-screen `WebView` and load `http://localhost:8990`.
- Enable JavaScript and DOM storage on the WebView.
- Override `onDestroy` to call `Mobile.stopServer()`.

---

## Verification Plan

### Manual Verification
1. Open the project in Android Studio or build it via Gradle CLI (`./gradlew assembleDebug`).
2. Deploy the APK to an Android Emulator or physical device.
3. Verify that the app opens directly into the dark-mode Glassmorphic Web UI.
4. Verify that the Android device broadcasts itself to the LAN and is visible on your PC's FileShare app.
5. Attempt a file transfer from the PC to the Android device and verify it lands in the Android Download folder without lag.
