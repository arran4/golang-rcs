#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="8311-rcsdiff-two-revs.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
Line 1
EOF
ci -q -t-'desc' -m'r1' -d'2020-01-01 00:00:00Z' file.txt
co -q -l file.txt
echo 'Line 2' >> file.txt
ci -q -m'r2' -d'2020-01-02 00:00:00Z' file.txt
co -q -l file.txt
echo 'Line 3' >> file.txt
# At this point: 1.1 has Line 1; 1.2 has Line 1+Line 2; Working has Line 1+Line 2+Line 3

# Run the command and capture output (allowing exit code 1 for diffs)
rcsdiff -r1.1 -r1.2 file.txt > output.txt 2>&1 || true

cat > "$OLDPWD/8311-rcsdiff-two-revs.txtar" <<EOF
-- description.txt --
rcsdiff -rREV1 -rREV2 (compare two revisions)

-- options.conf --
{"args": ["-r1.1", "-r1.2", "file.txt"] }

-- input.txt --
$(cat file.txt)

-- tests.txt --
rcsdiff

-- input.txt,v --
$(cat file.txt,v)

-- expected.out --
$(cat output.txt)
EOF
