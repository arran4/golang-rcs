#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="42-co-lock-rev-1.1.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "ONE\nTWO\n" > file.txt

# Create initial revision 1.1
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
# Disable strict locking
rcs -q -U file.txt
# Check out writable
co -q -l file.txt
# Modify file
echo "THREE" >> file.txt
# Check in as 1.2
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cp file.txt,v input.txt,v
rm file.txt

# Run co with -l1.1 to lock revision 1.1
co -q -l1.1 file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout revision 1.1 with lock

-- options.conf --
{"args": ["-q","-l1.1","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
ONE
TWO

-- expected.txt,v --
$(cat file.txt,v)
EOF
