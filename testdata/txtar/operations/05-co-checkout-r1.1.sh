# 05-co-checkout-r1.1.sh
#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="05-co-checkout-r1.1.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "ONE\nTWO\n" > file.txt

ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt
echo "THREE" >> file.txt
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cp file.txt,v input.txt,v
rm file.txt

co -q -r1.1 file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout revision 1.1

-- options.conf --
{"args": ["-q","-r1.1","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
$(cat file.txt)
EOF
