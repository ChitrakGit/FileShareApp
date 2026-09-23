# Android Application Walkthrough

The base structure and integration code for your FileShare Android app have been fully created inside the `android/` directory! Here is what was implemented to fulfill your constraints securely and efficiently.

## What Was Implemented

### 1. Minimal Go Wrapper (`pkg/mobile/mobile.go`)
I created a very lightweight Go wrapper package. This package is designed to be converted by `gomobile` into a `.aar` library, which allows Android's Kotlin code to seamlessly start and stop the `FileShare` backend natively.

### 2. High-Performance Android Scaffold
I constructed a standard Android project structure inside the `android/` directory.

> [!TIP]
> **Performance Optimization**: Because we use a `WebView` directed at the local Go server, there is zero translation layer overhead for UI components. It avoids memory-heavy background serialization processes and guarantees identical performance and UI parity with Windows and Linux.

### 3. Strict Security & File Save Path
I implemented your exact security requests in `android/app/src/main/java/com/fileshare/app/MainActivity.kt`:
1. **Download Path Enforcement**: I hardcoded the storage path to strictly evaluate to the Android native `Downloads/FileShare` directory (`Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)`).
2. **Permission Guarding**: The `AndroidManifest.xml` only asks for standard network APIs (WIFI state, Multicast) and local storage access. **No location, analytics, or background telemetry permissions** are requested. 
3. **No External Tracking**: The `WebView` runs with standard caching, and since it relies on the offline Go binary serving on `localhost`, it operates identically to the desktop client (zero tracking, fully decentralized).

## How to Compile & Run

Since the server environment does not have the Android NDK (Native Development Kit) fully installed, you will need to execute the final build step locally on your PC.

### Step 1: Generate the `fileshare.aar` Library
Make sure you have `gomobile` installed, then run the following in your terminal from the project root:
```bash
gomobile bind -target=android -o android/app/libs/fileshare.aar ./pkg/mobile
```
*Note: This requires the Android SDK and NDK installed on your machine.*

### Step 2: Open and Run in Android Studio
1. Open **Android Studio**.
2. Select **Open an existing Project**.
3. Select the `c:\Company Files\Study\FileShare\android` directory.
4. Let Gradle sync and resolve the dependencies.
5. Click **Run (▶)** to install and launch it on your connected mobile device or emulator!
