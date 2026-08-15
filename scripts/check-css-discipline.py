#!/usr/bin/env python3
"""Keeps the CSS split from silently undoing itself.

Two rules, both learned in this repository rather than borrowed:

  1. A selector in system.css has to be used by three or more files. Two is the
     moment to extract a component instead — promoting at two is how
     `.research-header`, `.tasks-header` and `.roadmaps-header` came to be three
     names for one rule.
  2. A `<style scoped>` block may not carry a hex colour, a box-shadow or an
     off-scale font-size. Once every component styles itself the tokens are the
     only thing holding the look together, so a literal in a scoped block is a
     piece of the design system escaping into one component.

Run: python3 scripts/check-css-discipline.py
"""
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FRONTEND = os.path.join(ROOT, "frontend")
SYSTEM = os.path.join(FRONTEND, "assets", "css", "system.css")

# Built as strings at runtime, so a static search finds nothing and would
# condemn them. See StatusBadge.vue and TagList.vue.
DYNAMIC = re.compile(r"^\.(badge|tag-hue|tag|hue)-")

# Deliberately global whatever the count: they style `v-html` output, or they
# are affordances that must not vary between the components that use them.
EXEMPT = {
    "container", "sr-only", "skip-link", "no-print", "markdown-content",
    "block-doc", "crossref-link", "search-mark", "user-avatar", "short-code",
    "skeleton-card", "empty-state", "page-title", "page-header", "card",
}

PRINT = re.compile(r"@media print")

# An escape hatch that has to say why. Syntax highlighting, a print palette and
# a full-bleed canvas are the three that legitimately need their own colours.
LITERAL_OK = re.compile(r"css-discipline:\s*literal-ok")


def sources():
    for base, dirs, files in os.walk(FRONTEND):
        dirs[:] = [d for d in dirs if d not in ("node_modules", ".output", ".nuxt", "dist")]
        for f in files:
            if f.endswith((".vue", ".ts")):
                path = os.path.join(base, f)
                with open(path, encoding="utf-8", errors="replace") as fh:
                    yield path, fh.read()


def scoped_blocks(src):
    """Only `<style scoped>`; a global block in a .vue file is a decision."""
    for m in re.finditer(r"<style[^>]*\bscoped\b[^>]*>(.*?)</style>", src, re.S):
        yield PRINT.split(m.group(1))[0]  # a palette for paper is literal on purpose


def main():
    issues = []
    files = dict(sources())

    with open(SYSTEM, encoding="utf-8") as fh:
        system = re.sub(r"/\*.*?\*/", "", fh.read(), flags=re.S)

    for sel in sorted(set(re.findall(r"^\.([a-z][\w-]*)", system, re.M))):
        if sel in EXEMPT or DYNAMIC.match("." + sel):
            continue
        users = {p for p, s in files.items() if re.search(r"[\"'\s.]%s[\"'\s]" % re.escape(sel), s)}
        if 0 < len(users) < 3:
            names = ", ".join(sorted(os.path.relpath(u, ROOT) for u in users))
            issues.append(
                ".%s is in system.css but only %d file(s) use it (%s). Three "
                "earn a place there; two means extract a component instead."
                % (sel, len(users), names)
            )

    for path, src in files.items():
        if not path.endswith(".vue"):
            continue
        rel = os.path.relpath(path, ROOT)
        for block in scoped_blocks(src):
            if LITERAL_OK.search(block):
                continue  # a documented palette: syntax highlighting, print, a canvas
            for lit in set(re.findall(r"#[0-9a-fA-F]{3,8}\b", block)):
                issues.append("%s: scoped style carries the literal %s — use a token." % (rel, lit))
            for m in re.finditer(r"box-shadow:([^;]+);", block):
                value = m.group(1).strip()
                # A lookahead here backtracks: `\s*` gives back the space and the
                # assertion then sees " var(", which is not "var(", so every
                # tokenised shadow was reported. Test the captured value instead.
                if value.startswith(("var(", "none", "inset")):
                    continue
                issues.append("%s: scoped box-shadow %s — use --shadow-1/2/3." % (rel, value[:40]))

    if issues:
        print("\n".join("- " + i for i in issues))
        print("\n%d issue(s)." % len(issues))
        return 1
    print("CSS discipline: clean.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
