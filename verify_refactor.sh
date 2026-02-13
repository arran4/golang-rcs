
echo "Testing gorcs commands with no args (via script to simulate TTY) (should print usage and exit with error)"

for cmd in format list-heads normalize-revisions validate to-json from-json; do
    echo "Testing $cmd..."
    script -q -c "go run ./cmd/gorcs $cmd" output_$cmd.log
    OUTPUT=$(cat output_$cmd.log)
    if echo "$OUTPUT" | grep -q "Usage:"; then
        echo "SUCCESS: Usage printed for $cmd"
    else
        echo "FAILURE: Usage NOT printed for $cmd"
        echo "Output was:"
        cat output_$cmd.log
    fi
    # Check exit code of the script/command
    # script's exit code is the child process's exit code
    # We expect non-zero because ensureFiles returns error
    script -q -c "go run ./cmd/gorcs $cmd" /dev/null
    if [ $? -ne 0 ]; then
        echo "SUCCESS: $cmd exited with error"
    else
         echo "FAILURE: $cmd exited with success (0)"
    fi
done

echo "Testing gorcs format with piped input (no args) (should NOT print usage)"
OUTPUT=$(echo "head 1.1; access; symbols; locks; comment @# @; 1.1 date 2021.01.01.00.00.00; author u; state Exp; branches; next; desc @@ 1.1 log @@ text @@" | go run ./cmd/gorcs format -s 2>&1)
EXIT_CODE=$?
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "FAILURE: Usage printed for piped format but shouldn't have"
else
    echo "SUCCESS: Usage NOT printed for piped format"
fi

# Clean up
rm output_*.log 2>/dev/null
