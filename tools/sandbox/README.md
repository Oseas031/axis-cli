# Axis Sandbox Image

Standard Linux environment for Agent execution. Includes Go toolchain and common unix utilities.

## Build

```bash
docker build -t axis-sandbox:latest tools/sandbox/
```

## Contents

- **Go 1.21** (compile, test, vet)
- **git** (version control)
- **curl, jq** (HTTP + JSON)
- **ripgrep (rg), fd** (fast search)
- **tree** (directory visualization)
- Standard coreutils (grep, find, sed, awk, wc, sort, etc.)

## Usage

`SandboxedBashTool` uses this image by default. Agent commands execute inside the container with the project directory mounted read-only at `/workspace`.

```go
tool := NewSandboxedBashTool(SandboxConfig{
    WorkDir: "/path/to/project",
    // Image defaults to "axis-sandbox:latest"
})
```
