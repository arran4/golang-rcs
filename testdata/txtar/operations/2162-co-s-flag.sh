#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="2162-co-s-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -sRel -t- -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

co -q -l file.txt
echo "New content" > file.txt
ci -q -sProd -mnew -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Now file.txt,v has 1.1 (Rel) and 1.2 (Prod)
# We want to checkout Rel (1.1)

cp file.txt,v input.txt,v
rm -f file.txt # Clean working dir

co -q -sRel file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co -s flag (checkout by state)

-- options.conf --
{"args": ["-sRel", "input.txt"]}

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# co

-- expected.txt --
$(cat file.txt)
EOF
