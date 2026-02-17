#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="3551-co-checkout-branch.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "MAIN\n" > file.txt
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

# Create a branch off 1.1
rcs -q -U file.txt
co -q -l1.1 file.txt
printf "BRANCH\n" > file.txt
ci -q -u -r1.1.1.1 -m"r1.1.1.1" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cp file.txt,v input.txt,v
rm file.txt

# Expected output is revision 1.1.1.1 content
printf "BRANCH\n" > expected_output.txt

# Check against system co
co -q -r1.1.1.1 -p file.txt > system_co_output.txt
diff expected_output.txt system_co_output.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout branch revision 1.1.1.1

-- options.conf --
{"args": ["-q","-r1.1.1.1","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
$(cat expected_output.txt)
EOF
