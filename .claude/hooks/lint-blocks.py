#!/usr/bin/env python3
"""PostToolUse gate keeping the block-document layers in step.

A block type only works when four places agree, and nothing at compile time ties
them together:

  1. the domain.BlockType constant        internal/domain/block.go
  2. a normalizer                          internal/service/blocks.go
  3. a text projection                     internal/service/blocks_text.go
  4. the catalog an agent reads            internal/docs/blocks.md

Miss the normalizer and the block is dropped on write. Miss the projection and its
prose stops being searchable and its [[E3]] references stop resolving. Miss the
catalog and an author guesses the field names — which is exactly how the sibling
project shipped a block that was valid, rendered, undocumented, and silently
discarded on every publish because everyone guessed the wrong key.

Measured when introduced: 0 hits on the 10 shipped block types.
"""

import json
import os
import re
import sys

DOMAIN = "internal/domain/block.go"
NORMALIZER = "internal/service/blocks.go"
PROJECTION = "internal/service/blocks_text.go"
CATALOG = "internal/docs/blocks.md"
RENDERER = "frontend/components/blocks/BlockRenderer.vue"

WATCHED = (DOMAIN, NORMALIZER, PROJECTION, CATALOG, RENDERER)


def read(root, rel):
    path = os.path.join(root, rel)
    try:
        return open(path, encoding="utf-8", errors="replace").read()
    except OSError:
        return None


def declared_types(src):
    """Block type string literals, wherever they are declared.

    Scanned across the whole file on purpose: block.go has several const blocks
    (caps, widths, variants), and anchoring to the first one found the caps and
    reported no types at all — which made this gate silently pass everything.
    """
    return set(re.findall(r'Block\w+\s+BlockType\s*=\s*"([a-z_]+)"', src))


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    path = (payload.get("tool_input") or {}).get("file_path", "")
    root = os.environ.get("CLAUDE_PROJECT_DIR", "")
    rel = os.path.relpath(path, root) if root and path else path
    if not any(rel.endswith(w) for w in WATCHED):
        return 0

    domain_src = read(root, DOMAIN)
    if not domain_src:
        return 0
    types = declared_types(domain_src)
    if not types:
        return 0

    norm_src = read(root, NORMALIZER) or ""
    proj_src = read(root, PROJECTION) or ""
    cat_src = read(root, CATALOG) or ""
    rend_src = read(root, RENDERER)

    issues = []
    for t in sorted(types):
        const = "Block" + "".join(p.capitalize() for p in t.split("_"))
        # HTML reads better than Html; accept the domain's own spelling too.
        consts = {const, "Block" + t.upper()}

        if not any(re.search(r"\b%s\b\s*:" % re.escape(c), norm_src) for c in consts):
            issues.append(
                f"block {t!r} has no normalizer in {NORMALIZER} — it will be "
                f"dropped on every write, silently."
            )
        if not any(re.search(r"\b%s\b" % re.escape(c), proj_src) for c in consts):
            issues.append(
                f"block {t!r} is not handled in {PROJECTION} — its text will not "
                f"be searchable and [[E3]] inside it will not resolve. Add it to "
                f"blockTextLines (explicitly return nil if it carries no prose) "
                f"and to BlockDocumentToMarkdown."
            )
        if not re.search(r"`%s`" % re.escape(t), cat_src):
            issues.append(
                f"block {t!r} is missing from the catalog in {CATALOG} — an agent "
                f"will guess the field names and the block will be discarded."
            )
        if rend_src is not None and t not in rend_src:
            issues.append(
                f"block {t!r} has no branch in {RENDERER} — it will store fine and "
                f"render as nothing."
            )

    if issues:
        print(
            json.dumps(
                {
                    "hookSpecificOutput": {
                        "hookEventName": "PostToolUse",
                        "additionalContext": "Block layers are out of step:\n"
                        + "\n".join("- " + i for i in issues),
                    }
                }
            )
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
