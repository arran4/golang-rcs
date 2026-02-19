# Test Operations Documentation

When configuring tests in `tests.txt` within `.txtar` files in this directory:

*   Use `rcs` as the command prefix (e.g., `rcs log`, `rcs diff`). This convention indicates an intermediate representation between the original RCS commands and the new `gorcs` commands.
*   Only specify the first subcommand (e.g., `log`) in `tests.txt`.
*   Do not include flags in `tests.txt`.
*   All flags and toggles should be defined in the `options.conf` or `options.json` file within the `.txtar` archive as JSON fields (e.g., `{"args": ["-sRel"]}`).
