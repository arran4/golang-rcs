# 02-ci-second-replace.sh
#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="02-ci-second-replace.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "ONE\nTWO\nTHREE\n" > file.txt

ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

cp file.txt input.txt
cp file.txt,v input.txt,v

printf "ALPHA\nBETA\nGAMMA\n" > file.txt

ci -q -u -m"replace-all" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci replace-all producing 1.2

-- options.conf --
{"args": ["-q","-u","-m","replace-all","-wtester","-d","2020-01-02 00:00:00Z","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
