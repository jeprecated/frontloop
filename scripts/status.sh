#!/usr/bin/env bash
set -euo pipefail

base="${1:-.}"
dir="$base/.frontloop"

if [ ! -d "$dir" ]; then
  echo "No .frontloop/ found at $base"
  echo "Run: ./scripts/init.sh $base"
  exit 1
fi

count() { find "$1" -maxdepth 1 -name '*.md' 2>/dev/null | wc -l; }

echo "=== Frontloop Status ==="
echo ""

# In progress
ip=$(count "$dir/in_progress")
if [ "$ip" -gt 0 ]; then
  echo "IN PROGRESS ($ip):"
  for f in "$dir"/in_progress/*.md; do
    title=$(grep '^title:' "$f" 2>/dev/null | head -1 | sed 's/^title: *//')
    echo "  * $title"
  done
  echo ""
fi

# Ready
r=$(count "$dir/ready")
echo "READY ($r):"
if [ "$r" -gt 0 ]; then
  for f in $(ls "$dir"/ready/*.md 2>/dev/null | sort); do
    title=$(grep '^title:' "$f" 2>/dev/null | head -1 | sed 's/^title: *//')
    pri=$(grep '^priority:' "$f" 2>/dev/null | head -1 | sed 's/^priority: *//')
    echo "  $(basename "$f" .md)  [$pri]  $title"
  done
else
  echo "  (empty)"
fi
echo ""

# Clarify
c=$(count "$dir/clarify")
echo "NEEDS CLARIFICATION ($c):"
if [ "$c" -gt 0 ]; then
  for f in "$dir"/clarify/*.md; do
    title=$(grep '^title:' "$f" 2>/dev/null | head -1 | sed 's/^title: *//')
    pri=$(grep '^priority:' "$f" 2>/dev/null | head -1 | sed 's/^priority: *//')
    echo "  $(basename "$f" .md)  [$pri]  $title"
  done
else
  echo "  (empty)"
fi
echo ""

# Done
d=$(count "$dir/done")
echo "DONE ($d):"
if [ "$d" -gt 0 ]; then
  for f in $(ls -t "$dir"/done/*.md 2>/dev/null | head -5); do
    title=$(grep '^title:' "$f" 2>/dev/null | head -1 | sed 's/^title: *//')
    echo "  $(basename "$f" .md)  $title"
  done
  [ "$d" -gt 5 ] && echo "  ... and $((d - 5)) more"
else
  echo "  (empty)"
fi
