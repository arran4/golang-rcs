#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="1492-rcsdiff-working-head.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
Line 1
Line 2
Line 3
EOF
ci -q -t-'desc' -m'r1' -d'2020-01-01 00:00:00Z' file.txt
co -q -l file.txt
# Modify line 2
sed -i 's/Line 2/Line 2 Modified/' file.txt

# Run the command and capture output
# 0 revisions: compares working file with latest revision (1.1)
rcsdiff file.txt > output.txt 2>&1 || true

cat > "$OLDPWD/1492-rcsdiff-working-head.txtar" <<EOF
-- description.txt --
rcsdiff no args (compare working file with head)

-- options.conf --
{"args": ["file.txt"] }

-- input.txt --
$(cat file.txt)

-- tests.txt --
rcsdiff

-- input.txt,v --
$(cat file.txt,v)

-- expected.out --
$(cat output.txt)
EOF
