# 01-ci-second-append.sh
#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/01-ci-second-append.txtar"
TMP_OUT="$OUTDIR/.01-ci-second-append.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "A\nB\n" > file.txt

# setup: create 1.1 and keep working file
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

cp file.txt   input.txt
cp file.txt,v input.txt,v

# mutate for 1.2
printf "C\nD\n" >> file.txt

# execution (ONE command)
ci -q -u -m"append" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cat > "$TMP_OUT" <<EOF
-- description.txt --
ci append change producing 1.2

-- options.conf --
{"args": ["-q","-u","-m","append","-wtester","-d","2020-01-02 00:00:00Z","input.txt"] }

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
