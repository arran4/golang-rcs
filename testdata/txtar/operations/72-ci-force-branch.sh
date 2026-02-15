#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAME="72-ci-force-branch"
OUT="$OUTDIR/$NAME.txtar"
TMP_OUT="$OUTDIR/.$NAME.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

# Create initial content
printf "INITIAL\nCONTENT\n" > file.txt

# setup: create 1.1
ci -q -i -u -m"initial" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Prepare input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# Modify file if needed
# ensure file is writable as ci -u makes it read-only
chmod +w file.txt
printf 'BRANCH\nCONTENT\n' > file.txt

# execution
# We run the command to generate the expected output
ci -q -u -f1.1.1.1 -m'forced-branch' -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cat > "$TMP_OUT" <<EOF
-- description.txt --
ci -f1.1.1.1 forces creation of branch 1.1.1

-- options.conf --
{"args": ["-q", "-u", "-f1.1.1.1", "-m", "forced-branch", "-wtester", "-d", "2020-01-02 00:00:00Z", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
ci

-- tests.md --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
