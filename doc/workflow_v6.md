# Version 6 Workflow

## Features / Tasks
1. Test every version (done via `go test ./test/... -v`).
2. Test that a full directory can be sent from one PC to another without lagging.

## Implementation Details
- We already have tar streaming logic in `pkg/transfer/archiver.go` which handles directories seamlessly.
- We will add an end-to-end test in `test/e2e_v6_test.go` that creates a directory structure with multiple files and subdirectories, then sends it using the existing QUIC client.
- We will verify that all files and subdirectories are reconstructed correctly on the receiver end.

## Testing
- The new test will be named `TestE2EV6_DirectoryTransfer`.
- It will verify the directory structure is preserved and transfer finishes without lagging.
