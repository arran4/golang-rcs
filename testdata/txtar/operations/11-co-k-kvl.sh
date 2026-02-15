#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="11-co-k-kvl.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo 'This is revision 1.1' > file.txt
echo '$Revision$' >> file.txt
echo '$Locker$' >> file.txt

# Create RCS file
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt
chmod u+w file.txt

# Lock revision 1.1
rcs -q -l1.1 file.txt

# Save RCS file
cp file.txt,v input.txt,v
rm file.txt

# Now run co with -kkvl
co -q -kkvl file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
co checkout -kkvl (with locker)

-- options.conf --
{"args": ["-q", "-kkvl", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# co

-- expected.txt --
$(cat file.txt)
EOF
