#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="10-co-k-kv.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo 'This is revision 1.1' > file.txt
echo '$Revision$' >> file.txt
echo '$Date$' >> file.txt
echo '$Author$' >> file.txt
echo '$State$' >> file.txt
echo '$Log$' >> file.txt
# echo '$Source$' >> file.txt # Excluded because it expands to absolute path
echo '$Id$' >> file.txt

# Create RCS file
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

# The file.txt now has expanded keywords because ci -u expands them.
# Let's save the RCS file as input.txt,v
cp file.txt,v input.txt,v
rm file.txt

# Now run co with -kkv (default) to generate expected output
co -q -kkv file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout -kkv (default)

-- options.conf --
{"args": ["-q", "-kkv", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# co

-- expected.txt --
$(cat file.txt)
EOF
