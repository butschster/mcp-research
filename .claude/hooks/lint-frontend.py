#!/usr/bin/env python3
"""PostToolUse gate for the Nuxt frontend of mcp-research.

Every check is derived from a real fix in this repository's history:

  1. bare-fetch    07d8666 "use authenticated fetch for all API calls".
                   A bare $fetch to /api/* goes out without a Bearer token and
                   gets a 401 as soon as auth_enabled is on.
  2. auto-import   ea31ebb "add explicit imports ... for Storybook".
                   A component relying on a Nuxt auto-import (renderRefs,
                   tagHue) crashes in Storybook, which has no auto-imports.
  3. missing-story every component under frontend/components has a .stories.ts
                   (100% coverage when this was introduced). A new component
                   without one drops out of the catalog.

Measured when introduced: 0 hits across 60 .vue files (the three $fetch calls in
settings.vue pass — they set Authorization by hand).
"""

import json
import os
import re
import sys

# Which helpers Storybook resolves for a component that did not import them.
#
# The list used to be two hand-maintained names while seven more had quietly
# joined the codebase, and `normalizeContent` was being called with no import in
# five components — a ReferenceError on render in the catalogue, not a warning.
# So the allow-list is no longer written here: it is read from the Storybook
# config that actually decides it, and anything a component calls which that
# config does not resolve is reported. The check now cannot drift from the
# thing it is checking.
STORYBOOK_CONFIG = "frontend/.storybook/main.ts"


def storybook_auto_imports(repo_root):
    """Symbol names `unplugin-auto-import` is configured to resolve."""
    cfg = os.path.join(repo_root, STORYBOOK_CONFIG)
    try:
        with open(cfg, encoding="utf-8") as fh:
            src = fh.read()
    except OSError:
        return None  # No config, no claim to make.
    block = re.search(r"AutoImport\(\{(.*?)\n\s*\}\)", src, re.S)
    if not block:
        return None
    return set(re.findall(r"'([A-Za-z_$][\w$]*)'", block.group(1)))



def project_helpers(repo_root):
    """Every function exported from composables/ and utils/, by name."""
    names = set()
    for sub in ("frontend/composables", "frontend/utils"):
        d = os.path.join(repo_root, sub)
        if not os.path.isdir(d):
            continue
        for entry in os.listdir(d):
            if not entry.endswith(".ts"):
                continue
            try:
                with open(os.path.join(d, entry), encoding="utf-8") as fh:
                    names.update(re.findall(r"export\s+function\s+([A-Za-z_$][\w$]*)", fh.read()))
            except OSError:
                pass
    return names


def check_bare_fetch(rel, lines):
    issues = []
    for i, line in enumerate(lines):
        if "$fetch" not in line or "authFetch" in line:
            continue
        stripped = line.strip()
        if stripped.startswith(("//", "*", "'", '"')):
            continue
        block = "\n".join(lines[i : i + 10])
        if "Authorization" in block:
            continue
        issues.append(
            f"{rel}:{i + 1}: bare $fetch to the API. Use authFetch() from "
            f"useAuth() or useApi(), or the request goes out without a Bearer token."
        )
    return issues


def check_auto_imports(rel, src, repo_root):
    """A component calling a project helper Storybook will not resolve for it.

    Nuxt auto-imports everything under composables/ and utils/; Storybook
    resolves only what its config names. The gap is a ReferenceError the moment
    the story renders, and it is invisible until somebody opens the catalogue.
    """
    allowed = storybook_auto_imports(repo_root)
    if allowed is None:
        return []
    issues = []
    for sym in sorted(project_helpers(repo_root)):
        if sym in allowed:
            continue
        if not re.search(r"\b%s\(" % sym, src):
            continue
        if re.search(r"(function|const|let)\s+%s\b" % sym, src):
            continue  # defined locally
        if re.search(r"import\s*\{[^}]*\b%s\b[^}]*\}\s*from" % sym, src):
            continue
        issues.append(
            f"{rel}: {sym}() is used without an explicit import, and "
            f"frontend/.storybook/main.ts does not resolve it either. Nuxt "
            f"auto-imports it, so this works in the app and throws a "
            f"ReferenceError in the catalogue. Add the import."
        )
    return issues


def check_story(path, rel):
    story = path[: -len(".vue")] + ".stories.ts"
    if os.path.isfile(story):
        return []
    return [
        f"{rel}: missing {os.path.basename(story)}. Every component under "
        f"frontend/components has a story — this one belongs in the catalog too."
    ]


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    path = (payload.get("tool_input") or {}).get("file_path", "")
    if not path.endswith((".vue", ".ts")) or ".stories." in path:
        return 0
    if "node_modules" in path or ".storybook" in path:
        return 0
    if not os.path.isfile(path):
        return 0

    root = os.environ.get("CLAUDE_PROJECT_DIR", "")
    rel = os.path.relpath(path, root) if root else path
    if not rel.startswith("frontend/"):
        return 0

    src = open(path, encoding="utf-8", errors="replace").read()
    lines = src.split("\n")

    issues = []
    if not rel.endswith("composables/useAuth.ts"):
        issues += check_bare_fetch(rel, lines)
    if rel.startswith("frontend/components/"):
        issues += check_auto_imports(rel, src, root)
        if rel.endswith(".vue"):
            issues += check_story(path, rel)

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
