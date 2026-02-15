/* gnu-h-v.c --- GNUish --help and --version handling

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

#include "base.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "gnu-h-v.h"
#include "b-complain.h"

int
nice_getopt (int argc, char **argv, const struct option *longopts)
{
  /* Support multiple calls.  */
  optind = 0;
  /* Don't display error message.  */
  opterr = 0;
  /* Do it!  */
  return getopt_long
    (argc, argv,
     "+",                       /* stop at first non-option */
     longopts, NULL);
}

#define COMMAND_VERSION                                         \
  (" (" PACKAGE_NAME ") " PACKAGE_VERSION "\n"                  \
   "Copyright (C) 2010-2020 Thien-Thi Nguyen\n"                 \
   "Copyright (C) 1990-1995 Paul Eggert\n"                      \
   "Copyright (C) 1982,1988,1989 Walter F. Tichy, Purdue CS\n"  \
   "License GPLv3+: GNU GPL version 3 or later"                 \
   " <http://gnu.org/licenses/gpl.html>\n"                      \
   "This is free software: you are free"                        \
   " to change and redistribute it.\n"                          \
   "There is NO WARRANTY, to the extent permitted by law.\n")

#define AB(blurb,uri)    blurb ": <" uri ">\n"
#define GNU(blurb,rest)  AB (blurb, "http://www.gnu.org/" rest)

#define BUGME                                           \
  ("\n"                                                 \
   AB ("Report bugs to", PACKAGE_BUGREPORT)             \
   GNU ("RCS home page", "software/rcs/")               \
   GNU ("General help using GNU software", "gethelp/"))

void
display_version (struct program const *prog, int flags)
{
  if (DV_WARN & flags)
    PWARN ("-V is obsolete; instead, use --version");
  printf ("%s%s", prog->name, COMMAND_VERSION);
  if (DV_EXIT & flags)
    exit (EXIT_SUCCESS);
}

enum hv_option_values
  {
    hv_help,
    hv_version
  };

static struct option const ok[] =
  {
    NICE_OPT ("help",    hv_help),
    NICE_OPT ("version", hv_version),
    NO_MORE_OPTIONS
  };

void
check_hv (int argc, char **argv, struct program const *prog)
{
  if (1 >= argc)
    return;

  switch (nice_getopt (argc, argv, ok))
    {
    case hv_help:
      {
        char usage[128];
        int nl;

        snprintf (usage, 128, "%s", prog->help);
        nl = strchr (usage, '\n') - usage;
        usage[nl] = '\0';

        printf ("Usage: %s %s\n\n%s\n%s%s",
                prog->name, usage,
                prog->desc,
                prog->help + nl,
                BUGME);
        exit (EXIT_SUCCESS);
      }
    case hv_version:
      display_version (prog, DV_EXIT);
    }
}

/* gnu-h-v.c ends here */
