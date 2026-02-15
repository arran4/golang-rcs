/* Clean up working files.

   Copyright (C) 2010-2020 Thien-Thi Nguyen
   Copyright (C) 1991, 1992, 1993, 1994, 1995 Paul Eggert

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
#include <errno.h>
#include <stdlib.h>
#include <dirent.h>
#include "same-inode.h"
#include "rcsclean.help"
#include "b-complain.h"
#include "b-divvy.h"
#include "b-esds.h"
#include "b-excwho.h"
#include "b-fb.h"
#include "b-feph.h"
#include "b-fro.h"

static void
cleanup (int *exitstatus, struct fro **workptr)
{
  if (FLOW (erroneous))
    *exitstatus = exit_failure;
  fro_zclose (&FLOW (from));
  fro_zclose (workptr);
  Ozclose (&FLOW (res));
  ORCSclose ();
  dirtempunlink ();
}

static bool
unlock (struct delta *delta)
{
  struct link box, *tp;

  if (delta && delta->lockedby
      && caller_login_p (delta->lockedby)
      && (box.next = GROK (locks))
      && (tp = lock_delta_memq (&box, delta)))
    {
      lock_drop (&box, tp);
      return true;
    }
  return false;
}

static int
valid_filename_p (const struct dirent *ent)
/* Filter out ".", "..", and comma-v filenames.
   This func is a predicate for ‘get_cwd_filenames’, q.v.  */
{
  const char *en = ent->d_name;

  if (en[0] == '.' && (!en[1] || (en[1] == '.' && !en[2])))
    return 0;

  if (rcssuffix (en))
    return 0;

  return 1;
}

static int
get_cwd_filenames (char ***aargv)
/* Put a vector of unsorted cwd's filenames into *AARGV.
   Ignore names of RCS files, as well as "." and "..".
   Allocate storage for the vector and entry names.
   Return the number of entries found.  */
{
  char dot[] = ".";
  struct dirent **names;
  int count, i;

  if (PROB (count = scandir (dot, &names, valid_filename_p, NULL)))
    fatal_sys (dot);

  *aargv = pointer_array (PLEXUS, count);
  for (i = count; i--;)
    {
      (*aargv)[i] = str_save (names[i]->d_name);
      free (names[i]);
    }
  free (names);

  return count;
}

DECLARE_PROGRAM (rcsclean, BOG_FULL);

static int
rcsclean_main (const char *cmd, int argc, char **argv)
{
  int exitstatus = EXIT_SUCCESS;
  struct fro *workptr = NULL;
  char *a, **newargv;
  char const *rev, *p;
  bool dounlock, perform, unlocked, unlockflag, waslocked, Ttimeflag;
  int expmode;
  struct wlink *deltas;
  struct delta *delta;
  struct stat workstat;

  CHECK_HV (cmd);
  gnurcs_init (&program);

  setrid ();

  expmode = -1;
  rev = NULL;
  perform = true;
  unlockflag = false;
  Ttimeflag = false;

  argc = getRCSINIT (argc, argv, &newargv);
  argv = newargv;
  for (;;)
    {
      if (--argc < 1)
        {
          argc = get_cwd_filenames (&newargv);
          argv = newargv;
          break;
        }
      a = *++argv;
      if (!*a || *a++ != '-')
        break;
      switch (*a++)
        {
        case 'k':
          if (0 <= expmode)
            redefined ('k');
          if (PROB (expmode = str2expmode (a)))
            goto unknown;
          break;

        case 'n':
          perform = false;
          goto handle_revision;

        case 'q':
          BE (quiet) = true;
          /* fall into */
        case 'r':
        handle_revision:
          chk_set_rev (&rev, a);
          break;

        case 'T':
          if (*a)
            goto unknown;
          Ttimeflag = true;
          break;

        case 'u':
          unlockflag = true;
          goto handle_revision;

        case 'V':
          setRCSversion (*argv);
          break;

        case 'x':
          BE (pe) = a;
          break;

        case 'z':
          zone_set (a);
          break;

        default:
        unknown:
          bad_option (*argv);
        }
    }

  dounlock = perform & unlockflag;

  if (FLOW (erroneous))
    cleanup (&exitstatus, &workptr);
  else
    for (; 0 < argc; cleanup (&exitstatus, &workptr), ++argv, --argc)
      {
        struct stat *repo_stat;
        char const *mani_filename;

        ffree ();

        if (!(0 < pairnames (argc, argv,
                             dounlock ? rcswriteopen : rcsreadopen,
                             true, true)
              && (mani_filename = MANI (filename))
              && (workptr = fro_open (mani_filename, FOPEN_R_WORK, &workstat))))
          continue;
        repo_stat = &REPO (stat);

        if (SAME_INODE (REPO (stat), workstat))
          {
            RERR ("RCS file is the same as working file %s.", mani_filename);
            continue;
          }

        p = NULL;
        if (rev)
          {
            struct cbuf numeric;

            if (!fully_numeric (&numeric, rev, workptr))
              continue;
            p = numeric.string;
          }
        else if (REPO (tip))
          switch (unlockflag ? findlock (false, &delta) : 0)
            {
            default:
              continue;
            case 0:
              p = GROK (branch) ? GROK (branch) : "";
              break;
            case 1:
              p = delta->num;
              break;
            }
        delta = NULL;
        deltas = NULL;                  /* Keep lint happy.  */
        if (p && !(delta = gr_revno (p, &deltas)))
          continue;

        waslocked = delta && delta->lockedby;
        BE (inclusive_of_Locker_in_Id_val) = unlock (delta);
        unlocked = BE (inclusive_of_Locker_in_Id_val) & unlockflag;
        if (unlocked < waslocked
            && workstat.st_mode & (S_IWUSR | S_IWGRP | S_IWOTH))
          continue;

        if (unlocked && !checkaccesslist ())
          continue;

        if (PROB (dorewrite (dounlock, unlocked)))
          continue;

        if (0 <= expmode)
          BE (kws) = expmode;
        else if (waslocked
                 && BE (kws) == kwsub_kv
                 && WORKMODE (repo_stat->st_mode, true) == workstat.st_mode)
          BE (kws) = kwsub_kvl;

        write_desc_maybe (FLOW (to));

        if (!delta
            ? workstat.st_size != 0
            : 0 < rcsfcmp (workptr, &workstat,
                           buildrevision (deltas, delta, NULL, false),
                           delta))
          continue;

        if (BE (quiet) < unlocked)
          aprintf (stdout, "rcs -u%s %s\n", delta->num, REPO (filename));

        if (perform & unlocked)
          {
            struct fro *from = FLOW (from);

            SAME_AFTER (from, delta->text);
            if (deltas->entry != delta)
              fro_trundling (true, from);
            if (PROB (donerewrite (true, file_mtime (Ttimeflag, repo_stat))))
              continue;
          }

        if (!BE (quiet))
          aprintf (stdout, "rm -f %s\n", mani_filename);
        fro_zclose (&workptr);
        if (perform && PROB (un_link (mani_filename)))
          syserror_errno (mani_filename);
      }

  tempunlink ();
  if (!BE (quiet))
    fclose (stdout);
  gnurcs_goodbye ();
  return exitstatus;
}

static const uint8_t rcsclean_aka[16] =
{
  2 /* count */,
  5,'c','l','e','a','n',
  8,'r','c','s','c','l','e','a','n'
};

YET_ANOTHER_COMMAND (rcsclean);

/*:help
[options] file ...
Options:
  -r[REV]       Specify revision.
  -u[REV]       Unlock if is locked and no differences found.
  -n[REV]       Dry run (no act, don't operate).
  -q[REV]       Quiet mode.
  -kSUBST       Substitute using mode SUBST (see co(1)).
  -T            Preserve the modification time on the RCS file
                even if it changes because a lock is removed.
  -V            Obsolete; do not use.
  -VN           Emulate RCS version N.
  -xSUFF        Specify SUFF as a slash-separated list of suffixes
                used to identify RCS file names.
  -zZONE        Specify date output format in keyword-substitution.

REV defaults to the latest revision on the default branch.
*/

/* rcsclean.c ends here */
