#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="15-co-k-v.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo 'This is revision 1.1' > file.txt
echo '$Revision$' >> file.txt
echo '$Date$' >> file.txt
echo '$Author$' >> file.txt

# Create RCS file
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

# file.txt has expanded keywords ($Revision: 1.1 $ etc)
cp file.txt,v input.txt,v
rm file.txt

# Now run co with -kv
co -q -kv file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout -kv (value only)

-- options.conf --
{"args": ["-q", "-kv", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# co

-- expected.txt --
$(cat file.txt)
EOF
