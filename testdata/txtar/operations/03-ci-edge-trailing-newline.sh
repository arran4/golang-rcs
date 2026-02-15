# 03-ci-edge-no-final-newline-change.sh
#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="03-ci-edge-no-final-newline-change.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "LINE1\nLINE2" > file.txt

ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

cp file.txt input.txt
cp file.txt,v input.txt,v

printf "LINE1\nLINE2X" > file.txt

ci -q -u -m"no-final-newline-change" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci edge case no trailing newline

-- options.conf --
{"args": ["-q","-u","-m","no-final-newline-change","-wtester","-d","2020-01-02 00:00:00Z","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
