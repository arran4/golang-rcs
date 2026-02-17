#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="82-ci-l.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1. Setup
printf 'Initial content\n' > file.txt
ci -q -i -t-desc -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

# 2. Save initial RCS state (input.txt,v)
# We want input.txt,v to be the state BEFORE the test command.
cp file.txt,v input.txt,v

# 3. Modify working file
printf "Locked content
" > file.txt

# 4. Save working file state (input.txt)
# We want input.txt to be the state BEFORE the test command (i.e. the modified content that will be checked in).
cp file.txt input.txt

# 5. Run the test command
ci -q -l '-mlocked checkin' -wtester '-d2020-01-02 00:00:00Z' file.txt </dev/null

# 6. Generate txtar
cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci checkin and lock

-- options.conf --
{"args": ["-q", "-l", "-mlocked checkin", "-wtester", "-d2020-01-02 00:00:00Z", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
