#!/usr/bin/env python3
"""PreToolUse gate on Bash commands.

The trap: 2916962 "exclude frontend/node_modules from go test paths".
`go test ./...` picks up Go files inside frontend/node_modules (the flatted
package) and fails. The Makefile and both CI workflows narrow the paths to
./cmd/... ./internal/... — a manual run should match.

Measured: fires only on the exact `go test ./...` shape; `go test ./internal/...`,
`make test` and `go build ./...` all pass.
"""

import json
import re
import sys


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    if payload.get("tool_name") != "Bash":
        return 0
    cmd = (payload.get("tool_input") or {}).get("command", "")

    if re.search(r"\bgo\s+test\b[^|&;]*\s\./\.\.\.", cmd):
        sys.stderr.write(
            "go test ./... picks up Go files from frontend/node_modules and fails.\n"
            "Use: make test  (or go test ./cmd/... ./internal/...)\n"
        )
        return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
