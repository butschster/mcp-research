#!/usr/bin/env bash
# Audit all Vue components and pages for Storybook story coverage.
# Run from project root: bash .claude/skills/storybook-setup/scripts/audit-components.sh

set -euo pipefail
cd "$(dirname "$0")/../../../../frontend"

echo "=== Component Storybook Audit ==="
echo "Date: $(date -Iseconds)"
echo ""

# Count total components
TOTAL=$(find components -name '*.vue' | wc -l | tr -d ' ')
echo "Total Vue components: $TOTAL"
echo ""

# Count stories
STORIES=$(find components -name '*.stories.ts' 2>/dev/null | wc -l | tr -d ' ')
echo "Story files: $STORIES"
echo "Coverage: $STORIES / $TOTAL"
echo ""

# Components WITHOUT stories
echo "=== Components WITHOUT stories ==="
while IFS= read -r vue_file; do
  story_file="${vue_file%.vue}.stories.ts"
  if [ ! -f "$story_file" ]; then
    echo "  MISSING: $vue_file"
  fi
done < <(find components -name '*.vue' | sort)
echo ""

# Components WITH stories
echo "=== Components WITH stories ==="
find components -name '*.stories.ts' -exec echo "  OK: {}" \; | sort
echo ""

# Component complexity (lines of code)
echo "=== Top 20 largest components (by LOC) ==="
find components -name '*.vue' -exec wc -l {} \; | sort -rn | head -20
echo ""

# Page complexity
echo "=== Page sizes (by LOC) ==="
find pages -name '*.vue' -exec wc -l {} \; | sort -rn
echo ""

# Components using defineProps
echo "=== Props summary ==="
grep -rn 'defineProps' components --include='*.vue' | sed 's/:.*//' | sort -u | while read -r f; do
  props=$(grep -A1 'defineProps' "$f" | grep -oP '\w+(?=[:?])' | tr '\n' ', ' | sed 's/,$//')
  echo "  $f: $props"
done
echo ""

# Components using composables
echo "=== Composable usage ==="
grep -rohP 'use[A-Z]\w+' components --include='*.vue' | sort | uniq -c | sort -rn | head -20
echo ""

# Components using renderRefs
echo "=== renderRefs usage ==="
grep -rn 'renderRefs' components pages --include='*.vue' | sed 's/:.*//' | sort -u
echo ""

echo "=== Done ==="
