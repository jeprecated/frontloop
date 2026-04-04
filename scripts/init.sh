#!/usr/bin/env bash
set -euo pipefail

base="${1:-.}"
dir="$base/.frontloop"

mkdir -p "$dir"/{ready,clarify,in_progress,done}
echo "Initialized frontloop at $dir/"
echo "  ready/       — tasks ready for the worker"
echo "  clarify/     — tasks needing human input"
echo "  in_progress/ — current worker task (max 1)"
echo "  done/        — completed tasks"
