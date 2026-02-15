#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="9201-rcsdiff-two-revs.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
Line A
EOF
ci -q -t-'desc' -m'r1' -d'2020-01-01 00:00:00Z' file.txt
co -q -l file.txt
echo 'Line B' >> file.txt
ci -q -m'r2' -d'2020-01-02 00:00:00Z' file.txt
co -q -l file.txt
echo 'Line C' >> file.txt

# Run the command and capture output
# 2 revisions: compares 1.1 with 1.2
rcsdiff -r1.1 -r1.2 file.txt > output.txt 2>&1 || true

cat > "$OLDPWD/9201-rcsdiff-two-revs.txtar" <<EOF
-- description.txt --
rcsdiff -r1.1 -r1.2 (compare 1.1 with 1.2)

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
