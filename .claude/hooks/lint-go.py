#!/usr/bin/env python3
"""PostToolUse gate for Go code in mcp-research.

Every check is derived from a real fix in this repository's history:

  1. nullable-input    f69e571 "make optional MCP tool input fields nullable".
                       An optional scalar field of an MCP tool input struct must
                       be a pointer, otherwise a client sending null gets
                       "type null, want string".
  2. tool-go-error     CLAUDE.md: a tool never returns a Go error. That is a
                       protocol error, not a call result.
  3. normalize-content 506999a "sanitize UTF-8 in MCP tool responses and service
                       inputs". Any user-supplied text reaching a service from a
                       request goes through normalizeContent.
  4. gofmt             the file must be formatted.

Measured when introduced (after fixing what it found): 0 hits across 34 files in
internal/mcp/tools and 12 in internal/service.
Before the fixes: 7 (nullable-input) + 12 (normalize-content) + 24 (gofmt).
"""

import json
import os
import re
import subprocess
import sys

# Single-line fields: a backslash is data here, so they need normalizeTitle.
# Expanding \n in a title splits it in two and breaks the markdown heading it
# becomes on export.
TITLE_FIELDS = "Title|Label|Name|DisplayName"
# Markdown bodies: MCP clients really do send escaped newlines, so these need
# normalizeContent, which expands them.
BODY_FIELDS = "Description|Content|Notes|Goal|Instruction|Text|Rationale|Answer|Focus"
REQ_VARS = "req|nr|er|qr|sec"


def check_mcp_tool(path, src):
    """Optional scalar fields of input structs must be pointers."""
    issues = []
    for m in re.finditer(r"type (\w+) struct \{(.*?)\n\}", src, re.S):
        struct, body = m.groups()
        for line in body.strip().split("\n"):
            fm = re.match(
                r"\s*(\w+)\s+([\w\.\*\[\]]+)\s+`json:\"([^\"]+)\""
                r"(?:\s+jsonschema:\"([^\"]*)\")?",
                line,
            )
            if not fm:
                continue
            field, typ, _, desc = fm.groups()
            desc = desc or ""
            if typ.startswith("*") or typ.startswith("[]"):
                continue
            # A field counts as optional when the handler treats it as a filter
            # (input.X != "") or the description marks it as not required.
            optional = re.search(r'input\.%s\s*!=\s*(""|0)' % field, src) or re.search(
                r"optional|filter by|default:", desc, re.I
            )
            if optional:
                issues.append(
                    f"{path}: {struct}.{field} is optional but declared as "
                    f"{typ}. Use a pointer (*{typ}) plus derefStr/derefFloat64, "
                    f"or a client sending null will fail schema validation."
                )
    for m in re.finditer(r"^\s*return\s+nil,\s*nil,\s*(err|fmt\.Errorf)", src, re.M):
        line = src[: m.start()].count("\n") + 1
        issues.append(
            f"{path}:{line}: tool returns a Go error. Use "
            f"errorResult()/validationErrorResult() instead — returning an error "
            f"is an MCP protocol failure, not a call result."
        )
    return issues


def check_service(path, src):
    """User text from a request must be normalized — with the right function."""
    issues = []
    for i, line in enumerate(src.split("\n"), 1):
        # The value may already be wrapped in a normalize* call, so allow an
        # optional wrapper between the field and the request variable.
        m = re.search(
            r"\b(%s|%s)\s*[:=]\s*(?:normalize\w+\()?\*?(%s)\."
            % (TITLE_FIELDS, BODY_FIELDS, REQ_VARS),
            line,
        )
        if not m:
            continue
        field = m.group(1)
        is_title = field in TITLE_FIELDS.split("|")
        wanted = "normalizeTitle" if is_title else "normalizeContent"
        if wanted in line:
            continue
        # The field may have been normalized earlier: req.Content = normalizeContent(...)
        if re.search(r"(%s)\.%s\s*=\s*%s" % (REQ_VARS, field, wanted), src):
            continue
        if wanted == "normalizeTitle":
            issues.append(
                f"{path}:{i}: {line.strip()[:70]} stores a single-line field "
                f"without normalizeTitle(). Using normalizeContent() here expands "
                f"a literal \\n and splits the value in two; leaving it raw lets "
                f"invalid UTF-8 into the database."
            )
        else:
            issues.append(
                f"{path}:{i}: {line.strip()[:70]} stores request text without "
                f"normalizeContent(). Literal \\n and invalid UTF-8 reach the database "
                f"and break the MCP response."
            )
    return issues


def check_gofmt(path):
    try:
        out = subprocess.run(
            ["gofmt", "-l", path], capture_output=True, text=True, timeout=10
        )
    except (FileNotFoundError, subprocess.SubprocessError):
        return []
    if out.stdout.strip():
        return [f"{path}: not gofmt-clean. Run: gofmt -w {path}"]
    return []


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    path = (payload.get("tool_input") or {}).get("file_path", "")
    if not path.endswith(".go") or path.endswith("_test.go"):
        return 0

    root = os.environ.get("CLAUDE_PROJECT_DIR", "")
    rel = os.path.relpath(path, root) if root else path
    if not os.path.isfile(path):
        return 0

    src = open(path, encoding="utf-8", errors="replace").read()

    issues = check_gofmt(path)
    if rel.startswith("internal/mcp/tools/"):
        issues += check_mcp_tool(rel, src)
    if rel.startswith("internal/service/") and rel.endswith("_service.go"):
        if "export_service.go" not in rel:
            issues += check_service(rel, src)

    if issues:
        print(
            json.dumps(
                {
                    "hookSpecificOutput": {
                        "hookEventName": "PostToolUse",
                        "additionalContext": "Known traps in this repository:\n"
                        + "\n".join("- " + i for i in issues),
                    }
                }
            )
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
