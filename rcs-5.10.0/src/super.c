/* Dispatch an RCS command.

   Copyright (C) 2013-2020 Thien-Thi Nguyen

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
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include "super.help"
#include "b-divvy.h"
#include "b-complain.h"
#include "b-peer.h"

/* {Dynamic Root} TODO: Move into library.

   For the present, internal dispatch (e.g., ‘diff’ calls ‘co’) goes
   through execve(2), but the plan is to eventually elide that into
   a function call, at which point dynamic-root support will need to
   move into the library.  */

struct dynamic_root
{
  struct top *top;
  struct divvy *single;
  struct divvy *plexus;
  /* FIXME: What about these?
     - program_invocation_name
     - program_invocation_short_name
     - stderr
     - stdin
     - stdout
     (These are from "nm --defined-only -D rcs".)  */
};

static void
droot_global_to_stack (struct dynamic_root *dr)
{
  dr->top = top;
  dr->single = single;
  dr->plexus = plexus;
}

static void
droot_stack_to_global (struct dynamic_root *dr)
{
  top = dr->top;
  single = dr->single;
  plexus = dr->plexus;
}

static void
dispatch (int *exitval, submain_t *sub,
          const char *cmd, int argc, char **argv)
{
  struct dynamic_root super;

  droot_global_to_stack (&super);
  *exitval = sub (cmd, argc, argv);
  droot_stack_to_global (&super);
}

#define DECLARE_YA(prog)  extern const struct yacmd YA (prog)

DECLARE_YA (ci);
DECLARE_YA (co);
DECLARE_YA (rcs);
DECLARE_YA (rcsclean);
DECLARE_YA (rcsdiff);
DECLARE_YA (rcsmerge);
DECLARE_YA (rlog);

#define AVAIL(prog)  & YA (prog)

static const struct yacmd *avail[] =
  {
    AVAIL (ci),
    AVAIL (co),
    AVAIL (rcs),
    AVAIL (rcsclean),
    AVAIL (rcsdiff),
    AVAIL (rcsmerge),
    AVAIL (rlog)
  };

static const size_t n_avail = sizeof (avail) / sizeof (avail[0]);

static submain_t *
recognize (const char *maybe)
{
  size_t mlen = strlen (maybe);
  size_t i;

  for (i = 0; i < n_avail; i++)
    {
      const struct yacmd *y = avail[i];
      const uint8_t *aka = y->aka;
      size_t count = *aka++;

      while (count--)
        {
          struct tinysym *sym = (struct tinysym *) aka;

          if (mlen == sym->len
              && looking_at (sym, maybe))
            return y->func;
          aka += sym->len + 1;
        }
    }
  return NULL;
}

/* (length "who-groks-life-the-universe-and-everything")
   => 42
   :-D  */
#define MAX_COMMAND_SIZE  64

/* FIXME: This should probably be part of the ‘struct cbuf’
   group of procs, exposed in base.h.  */
static void
string_from_sym (char dest[MAX_COMMAND_SIZE],
                 const struct tinysym *sym)
{
  size_t len = (MAX_COMMAND_SIZE > sym->len
                ? sym->len
                : MAX_COMMAND_SIZE - 1);
  char *end = mempcpy (dest, sym->bytes, len);

  *end = '\0';
}

static void
display_commands (void)
{
  printf ("%-10s  %s\n", "(command)", "(description)");
  for (size_t i = 0; i < n_avail; i++)
    {
      const struct yacmd *y = avail[i];
      const uint8_t *aka = y->aka;
      char name[MAX_COMMAND_SIZE];

      string_from_sym (name, (struct tinysym *) (++aka));
      printf (" %-10s  %s\n", name, y->pr->desc);
    }
}

static void
display_aliases (void)
{
  printf ("%-10s  %s\n", "(command)", "(aliases)");
  for (size_t i = 0; i < n_avail; i++)
    {
      const struct yacmd *y = avail[i];
      const uint8_t *aka = y->aka;
      size_t count = *aka++;

      for (size_t j = 0; j < count; j++)
        {
          struct tinysym *sym = (struct tinysym *) aka;
          char name[MAX_COMMAND_SIZE];

          string_from_sym (name, sym);
          switch (j)
            {
            case 0:
              printf (" %-10s ", name);
              break;
            case 1:
              printf (" %s", name);
              break;
            default:
              printf (", %s", name);
            }
          aka += 1 + sym->len;
        }
      printf ("\n");
    }
}

struct top *top;

static bool
all_options_short_p (char **argv)
{
  bool ok;
  int i;

  for (ok = true, i = 1; argv[i]; i++)
    {
      if ('-' != argv[i][0])
        break;
      if ('-' == argv[i][1])
        {
          ok = false;
          break;
        }
    }
  return ok;
}

static char const hint[] = " (try --help)";

static exiting void
huh (const char *what, const char *argv1)
{
  PFATAL ("unknown %s: %s%s", what, argv1, hint);
}

#define HUH(what)  huh (what, argv[1])

DECLARE_PROGRAM (super, TYAG_IMMEDIATE);

int
main (int argc, char *argv[VLA_ELEMS (argc)])
{
  const char *cmd;
  submain_t *sub;
  int exitval = EXIT_SUCCESS;           /* momentary optimism */

  /* Reorder invocation ‘--help COMMAND’ to be ‘COMMAND --help’.
     FIXME (maybe): This behaves weirdly for ‘--help FILENAME’.  */
  if (3 == argc
      && STR_SAME ("--help", argv[1]))
    {
      char *tmp = argv[2];

      argv[2] = argv[1];
      argv[1] = tmp;
    }

  CHECK_HV (peer_super.meaningful);
  gnurcs_init (&program);

  /* Try the legacy interface first.  */
  if (2 > argc
      || ('-' == argv[1][0]
          && all_options_short_p (argv)))
    {
    legacy:
      sub = recognize (cmd = "rcs");
      dispatch (&exitval, sub, cmd, argc, argv);
    }
  else
    {
      /* Option processing.  */
      if ('-' == argv[1][0])
        {
          /* "DDC" stands for "dash-dash command".  */
          enum ddc_option_values
          {
            ddc_unrecognized = 0,
            ddc_commands,
            ddc_aliases
          };
          struct option allddc[] =
            {
              NICE_OPT ("commands", ddc_commands),
              NICE_OPT ("aliases",  ddc_aliases),
              NO_MORE_OPTIONS
            };

          switch (nice_getopt (argc, argv, allddc))
            {
            case ddc_commands:
              display_commands ();
              goto done;
            case ddc_aliases:
              display_aliases ();
              goto done;
            }
          /* No luck, sorry.  */
          HUH ("option");
        }

      /* Try dispatch.  */
      if ((sub = recognize (cmd = argv[1])))
        {
          /* Construct a simulated invocation.  */
          argv[1] = one_beyond_last_dir_sep (argv[0])
            ? argv[0]
            : str_save (PEER_SUPER ());

          dispatch (&exitval, sub, cmd, argc - 1, argv + 1);
        }

      /* Maybe support backward compatibility w/ obsolescent usage (ugh).
         This is "maybe" because it's just a heuristic (double-ugh).
         On the positive side, we do this only after trying to recognize
         the command (the common case), so the impact is reduced.

         TODO: Move "usage is obsolescent" handling from rcs.c to here,
         zonk the heuristic, and be happy (maybe).  */
      else if (strchr (cmd, SLASH)
               || ! PROB (access (cmd, R_OK)))
        goto legacy;

      /* No luck, sorry.  */
      else
        HUH ("command");
    }

 done:
  gnurcs_goodbye ();
  return exitval;
}

/*:help
[options] command [command-arg ...]
Options:
  --commands       Display available commands and exit.
  --aliases        Display command aliases and exit.
  --help COMMAND   Display help for COMMAND.

To display help for the legacy interface, use:
  --help frob
*/

/* super.c ends here */
