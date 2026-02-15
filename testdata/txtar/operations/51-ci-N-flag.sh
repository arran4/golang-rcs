#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="51-ci-N-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
Initial content.
EOF

# Initial checkin with symbol TAG pointing to 1.1
ci -q -i -u -m"r1" -nTAG -wtester -d'2021-01-01 00:00:00Z' file.txt </dev/null

# Prepare input state (1.1 checked in, TAG=1.1)
# Make file writable to modify it, simulating work
rcs -q -U file.txt
chmod u+w file.txt
cp file.txt,v input.txt,v

# Modify file for 1.2
cat >> file.txt <<'EOF'
Modified content.
EOF
cp file.txt input.txt

# Execute ci with -NTAG (should move TAG to 1.2)
ci -q -u -m"r2" -NTAG -wtester -d'2021-01-02 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci with -N flag (overwrite symbolic name)

-- options.conf --
{"args": ["-q","-u","-m","r2","-NTAG","-wtester","-d","2021-01-02 00:00:00Z","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
