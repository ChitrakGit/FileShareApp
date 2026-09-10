# FileShare Project Rules

When working on this project, the agent must adhere to the following rules:

1. **Documentation Location**: 
   - All project-related documentation (e.g., test reports, workflow documents, architecture diagrams, user manuals) must be generated and saved exclusively in the `doc/` directory. 
   - Do not create or save documentation in the root directory or ad-hoc folders like `test/`. 

2. **Build Artifacts Location**:
   - All compiled binaries, executables (e.g., `*.exe`), and build outputs must be generated and saved inside the `builds/` directory.
   - When running build commands, explicitly output to the `builds/` directory (e.g., `go build -o builds/fileshare.exe ./cmd/fileshare`).
   - Create the `builds/` directory if it does not already exist before building.

3. **Version Work & Workflows**:
   - When given work for a new version (e.g., V3, V4), **first** create a workflow document and save it in the `doc/` directory with the version included in the filename (e.g., `doc/workflow_v3.md`).
   - Proceed with implementing the features according to the approved workflow.
   - Generate test files corresponding to the new version's features and save them in the `test/` directory. The test files must also include the version in their filenames (e.g., `test/e2e_v3_test.go`).
   - After successfully running and passing the test cases, you must update the `README.md` file to reflect the new version's features and commands.
