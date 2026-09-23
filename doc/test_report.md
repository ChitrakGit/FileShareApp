# FileShare Test Report
Generated on: Wed, 23 Sep 2026 16:30:08 UTC

## Test Results
| Package | Test Name | Status | Duration |
|---------|-----------|--------|----------|
| `fileshare/test/android` | `TestAndroidModule` | ✅ PASS | 0.000s |
| `fileshare/test/linux` | `TestLinuxStartup` | ✅ PASS | 0.000s |
| `fileshare/test/cli` | `TestDiscoveryModule` | ✅ PASS | 0.000s |
| `fileshare/test/cli` | `TestTransferModule` | ✅ PASS | 0.000s |
| `fileshare/test/windows` | `TestWindowsModule` | ✅ PASS | 0.000s |
| `fileshare/test` | `TestEndToEnd/API_Info_Endpoint` | ✅ PASS | 0.000s |
| `fileshare/test` | `TestEndToEnd/File_Transfer_Streaming` | ✅ PASS | 0.110s |
| `fileshare/test` | `TestEndToEnd` | ✅ PASS | 1.110s |
| `fileshare/test` | `TestEndToEndV3/CrossPlatform_DefaultDownloadDir` | ✅ PASS | 0.000s |
| `fileshare/test` | `TestEndToEndV3` | ✅ PASS | 0.000s |
| `fileshare/test` | `TestE2EV4_QUIC_Transfer` | ✅ PASS | 1.610s |
| `fileshare/test` | `TestE2EV5_AddDeviceOnTheFly` | ✅ PASS | 0.000s |
| `fileshare/test` | `TestE2EV6_DirectoryTransfer` | ✅ PASS | 1.610s |

## Test Logs
```text
=== RUN   TestAndroidModule
    android_test.go:9: Android mobile module functions StartServer and StopServer are verified for structure.
--- PASS: TestAndroidModule (0.00s)
PASS
ok  	fileshare/test/android	0.004s
=== RUN   TestLinuxStartup
    linux_test.go:9: Linux-specific startup logic tested. Note: AppIndicator dependencies omitted from test environment.
--- PASS: TestLinuxStartup (0.00s)
PASS
ok  	fileshare/test/linux	0.004s
=== RUN   TestEndToEnd
[Server] FileShare listening on http://192.168.0.2:9099
[Server] Destination folder: /tmp/TestEndToEnd4122017281/001
=== RUN   TestDiscoveryModule
--- PASS: TestDiscoveryModule (0.00s)
=== RUN   TestTransferModule
    cli_test.go:18: CLI Transfer module tested via QUIC capabilities.
--- PASS: TestTransferModule (0.00s)
PASS
ok  	fileshare/test/cli	0.009s
=== RUN   TestWindowsModule
    windows_test.go:9: Windows-specific code tested. Note: cgo/Wails components omitted from direct test execution due to missing C libraries on the test environment.
--- PASS: TestWindowsModule (0.00s)
PASS
ok  	fileshare/test/windows	0.004s
2026/09/23 16:30:04 failed to sufficiently increase receive buffer size (was: 208 kiB, wanted: 7168 kiB, got: 416 kiB). See https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes for details.
[Server] QUIC Transport listening on UDP port 9099
=== RUN   TestEndToEnd/API_Info_Endpoint
    e2e_test.go:57: Successfully retrieved MAC Address: 06:00:c0:a8:00:02
--- PASS: TestEndToEnd/API_Info_Endpoint (0.00s)
=== RUN   TestEndToEnd/File_Transfer_Streaming

[Sender] Preparing to send 1 items (17 B) to 127.0.0.1:9099...
[Receiver] Checksum verification passed for 1 files.

[QUIC] Successfully unpacked 1 files (17 B) into /tmp/TestEndToEnd4122017281/001

[Sender] Warning: did not receive clean ACK from server: Application error 0x0 (remote)
[K[██████████████████████████████] 100.0% | 17 B/17 B | 0.0 MB/s | ETA: 0s
[Sender] Transfer completed successfully over QUIC in 107ms!

    e2e_test.go:84: Successfully sent and verified received file content!
--- PASS: TestEndToEnd/File_Transfer_Streaming (0.11s)
--- PASS: TestEndToEnd (1.11s)
=== RUN   TestEndToEndV3
=== RUN   TestEndToEndV3/CrossPlatform_DefaultDownloadDir
    e2e_v3_test.go:36: Successfully resolved OS-specific path: /home/jules/Downloads/FileShare
--- PASS: TestEndToEndV3/CrossPlatform_DefaultDownloadDir (0.00s)
--- PASS: TestEndToEndV3 (0.00s)
=== RUN   TestE2EV4_QUIC_Transfer
[Server] FileShare listening on http://192.168.0.2:8995
[Server] Destination folder: /tmp/fileshare-v4-download3624481098
[Server] QUIC Transport listening on UDP port 8995

[Sender] Preparing to send 1 items (29 B) to 127.0.0.1:8995...
[Receiver] Checksum verification passed for 1 files.

[QUIC] Successfully unpacked 1 files (29 B) into /tmp/fileshare-v4-download3624481098

[Sender] Warning: did not receive clean ACK from server: Application error 0x0 (remote)
[K[██████████████████████████████] 100.0% | 29 B/29 B | 0.0 MB/s | ETA: 0s
[Sender] Transfer completed successfully over QUIC in 106ms!

--- PASS: TestE2EV4_QUIC_Transfer (1.61s)
=== RUN   TestE2EV5_AddDeviceOnTheFly
--- PASS: TestE2EV5_AddDeviceOnTheFly (0.00s)
=== RUN   TestE2EV6_DirectoryTransfer
[Server] FileShare listening on http://192.168.0.2:8996
[Server] Destination folder: /tmp/fileshare-v6-download1960117275
[Server] QUIC Transport listening on UDP port 8996

[Sender] Preparing to send 4 items (42 B) to 127.0.0.1:8996...
[Receiver] Checksum verification passed for 2 files.

[QUIC] Successfully unpacked 2 files (42 B) into /tmp/fileshare-v6-download1960117275

[Sender] Warning: did not receive clean ACK from server: Application error 0x0 (remote)
[K[██████████████████████████████] 100.0% | 42 B/42 B | 0.0 MB/s | ETA: 0s
[Sender] Transfer completed successfully over QUIC in 107ms!

    e2e_v6_test.go:75: Directory transferred in 106.780992ms
--- PASS: TestE2EV6_DirectoryTransfer (1.61s)
PASS
ok  	fileshare/test	4.341s
```
