/* gnu-h-v.h --- GNUish --help and --version handling

   Copyright (C) 2010-2020 Thien-Thi Nguyen

   This file is part of GNU RCS.

   GNU RCS is free software: you can redistribute it and/or modify it
   under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   GNU RCS is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty
   of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
   See the GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

#include <getopt.h>

/* Clear ‘optind’ and ‘opterr’ then call ‘getopt_long’, arranging
   to do not permute ‘argv’.  Return what ‘getopt_long’ returns.  */
extern int
nice_getopt (int argc, char **argv, const struct option *longopts)
  ALL_NONNULL;

#define DV_ONLY   0
#define DV_WARN   1
#define DV_EXIT   2

/* Display the version blurb to stdout, starting with:
   | NAME (GNU RCS) PACKAGE_VERSION
   | ...
   and ending with newline.  NAME is the value of ‘prog->name’.
   FLAGS is the logical-OR of:
   | DV_ONLY -- don't do anything special
   | DV_WARN -- warn that this usage is obsolete (for ‘-V’);
   |            suggest using --version, instead
   | DV_EXIT -- finish w/ ‘exit (EXIT_SUCCESS)’
   The default is 0.  */
extern void
display_version (struct program const *prog, int flags)
  ALL_NONNULL;

/* If ARGC is less than 2, do nothing.
   If ARGV[1] is "--version", use ‘display_version’ and exit successfully.
   If ARGV[1] is "--help", display the help blurb, starting with:
   | NAME HELP
   and exit successfully.  NAME is the value of ‘prog->name’,
   while HELP is the value of ‘prog->help’.  */
extern void
check_hv (int argc, char **argv, struct program const *prog)
  ALL_NONNULL;

/* Idioms.  */

#define NICE_OPT(name,value)  \
  { name, no_argument, NULL, value }

#define NO_MORE_OPTIONS \
  {NULL, 0, NULL, 0}

#define CHECK_HV(cmd)  do                       \
    {                                           \
      program.invoke = argv[0];                 \
      program.name = cmd;                       \
      check_hv (argc, argv, &program);          \
    }                                           \
  while (0)

#define DECLARE_PROGRAM(prog,__tyag)            \
  static struct program program =               \
    {                                           \
      .desc = prog ## _blurb,                   \
      .help = prog ## _help,                    \
      .tyag = __tyag                            \
    }

/* gnu-h-v.h ends here */
