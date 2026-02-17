#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="14-co-k-b.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo 'This is binary file' > file.txt
echo '$Revision$' >> file.txt

# Initialize RCS file as binary
rcs -i -kb file.txt

# Check in
ci -q -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
# Since it is binary, keywords should NOT be expanded.
# file.txt should still contain $Revision$
# And RCS file should store $Revision$ (unexpanded).

cp file.txt,v input.txt,v
rm file.txt

# Now run co with -kb
co -q -kb file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout -kb (binary)

-- options.conf --
{"args": ["-q", "-kb", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# co

-- expected.txt --
$(cat file.txt)
EOF
