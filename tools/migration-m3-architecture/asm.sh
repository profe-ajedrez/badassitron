#!/usr/bin/env bash
# Historical two-build ASM comparison was archived in the M3-F source checkpoint.
set -euo pipefail
printf 'Historical rollback sources are not shipped in M7-A. Use the archived M3-F snapshot to reproduce ASM comparisons; use tools/release/bench-backend.sh for current performance.\n' >&2
exit 2
