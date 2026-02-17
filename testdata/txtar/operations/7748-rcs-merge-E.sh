#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="7748-rcs-merge-E.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# Initial file
printf "A\nB\nC\n" > file.txt
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

# Revision 1.2
co -q -l file.txt
printf "A\nB-changed\nC\n" > file.txt
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Prepare working file for merge
co -q -r1.1 file.txt
printf "A\nB-local\nC\n" > file.txt

# Save state
cp file.txt input.txt
cp file.txt,v input.txt,v

# Merge 1.1 -> 1.2 into working file with -E
rcsmerge -q -E -r1.1 -r1.2 file.txt || true

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs merge -E conflict

-- options.conf --
{"args": ["-q","-E","-r1.1","-r1.2","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs merge

-- expected.txt --
$(cat file.txt)
EOF
