#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="5830-rcsdiff-one-rev.txtar"
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

# Run the command and capture output
# 1 revision: compares working file with 1.1
rcsdiff -r1.1 file.txt > output.txt 2>&1 || true

cat > "$OLDPWD/5830-rcsdiff-one-rev.txtar" <<EOF
-- description.txt --
rcsdiff -r1.1 (compare working file with 1.1)

-- options.conf --
{"args": ["-r1.1", "file.txt"] }

-- input.txt --
$(cat file.txt)

-- tests.txt --
rcsdiff

-- input.txt,v --
$(cat file.txt,v)

-- expected.out --
$(cat output.txt)
EOF
