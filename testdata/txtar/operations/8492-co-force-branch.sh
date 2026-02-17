#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/8492-co-force-branch.txtar"
TMP_OUT="$OUTDIR/.8492-co-force-branch.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "v1\n" > file.txt
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Create 1.2
co -q -l file.txt
printf "v2\n" > file.txt
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Create branch 1.2.1.1
# Need to checkout locked first.
rm -f file.txt
co -q -l file.txt
printf "v2branch\n" > file.txt
ci -q -u -r1.2.1.1 -m"r2branch" -wtester -d'2020-01-03 00:00:00Z' file.txt </dev/null

# Now restore working file to HEAD (1.2) by checking out 1.2 explicitly
rm -f file.txt
co -q -r1.2 file.txt
# modify working file
chmod u+w file.txt
printf "modified\n" > file.txt

# save input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: -f1.2.1.1 forces overwrite with branch revision
co -q -f1.2.1.1 file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
co -f<BRANCH> forces overwrite of modified working file with branch revision

-- options.conf --
{"args": ["-q","-f1.2.1.1","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
$(cat file.txt)
EOF

mv -f "$TMP_OUT" "$OUT"
