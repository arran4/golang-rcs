/* Check out working files from revisions of RCS files.

   Copyright (C) 2010-2020 Thien-Thi Nguyen
   Copyright (C) 1990, 1991, 1992, 1993, 1994, 1995 Paul Eggert
   Copyright (C) 1982, 1988, 1989 Walter Tichy

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
#include "same-inode.h"
#include "timespec.h"
#include "co.help"
#include "b-complain.h"
#include "b-divvy.h"
#include "b-esds.h"
#include "b-excwho.h"
#include "b-fb.h"
#include "b-feph.h"
#include "b-fro.h"
#include "b-isr.h"
#include "b-peer.h"

static const char ks_hws[] = " \t";

#define SKIP_HWS(var)  var += strspn (var, ks_hws)

struct work
{
  struct stat st;
  bool force;
};

static char const quietarg[] = "-q";

/* State for -j.  */
struct jstuff
{
  struct divvy *jstuff;
  struct link head, *tp;
  struct symdef *merge;
  char const *expand, *suffix, *version, *zone;

  struct delta *d;
  /* Final delta to be generated.  */

  char const **ls;
  /* Revisions to be joined.  */

  int lastidx;
  /* Index of last element in `ls'.  */
};

static void
cleanup (int *exitstatus, FILE **neworkptr)
{
  FILE *mstdout = MANI (standard_output);

  if (FLOW (erroneous))
    *exitstatus = exit_failure;
  fro_zclose (&FLOW (from));
  ORCSclose ();
  if (FLOW (from)
      && STDIO_P (FLOW (from))
      && FLOW (res)
      && FLOW (res) != mstdout)
    Ozclose (&FLOW (res));
  if (*neworkptr != mstdout)
    Ozclose (neworkptr);
  dirtempunlink ();
}

static bool
rmworkfile (struct work *work)
/* Prepare to remove ‘MANI (filename)’, if it exists, and if it is read-only.
   Otherwise (file writable), if !quietmode, ask the user whether to
   really delete it (default: fail); otherwise fail.
   Return true if permission is gotten.  */
{
  if (work->st.st_mode & (S_IWUSR | S_IWGRP | S_IWOTH) && !work->force)
    {
      char const *mani_filename = MANI (filename);

      /* File is writable.  */
      if (!yesorno (false, "writable %s exists%s; remove it",
                    mani_filename, (stat_mine_p (&work->st)
                                    ? ""
                                    : ", and you do not own it")))
        {
          if (!BE (quiet) && ttystdin ())
            PERR ("checkout aborted");
          else
            PERR ("writable %s exists; checkout aborted", mani_filename);
          return false;
        }
    }
  /* Actual unlink is done later by caller.  */
  return true;
}

static int
rmlock (struct delta const *delta)
/* Remove the lock held by caller on ‘delta’.  Return -1 if
  someone else holds the lock, 0 if there is no lock on delta,
  and 1 if a lock was found and removed.  */
{
  struct link box, *tp;
  struct rcslock const *rl;

  box.next = GROK (locks);
  if (! (tp = lock_delta_memq (&box, delta)))
    /* No lock on ‘delta’.  */
    return 0;
  rl = tp->next->entry;
  if (!caller_login_p (rl->login))
    /* Found a lock on ‘delta’ by someone else.  */
    {
      RERR ("revision %s locked by %s; use co -r or rcs -u",
            delta->num, rl->login);
      return -1;
    }
  /* Found a lock on ‘delta’ by caller; delete it.  */
  lock_drop (&box, tp);
  return 1;
}

static void
jpush (char const *rev, struct jstuff *js)
{
  js->tp = extend (js->tp, rev, js->jstuff);
  js->lastidx++;
}

static char *
addjoin (char *spec, struct jstuff *js)
/* Tokenize ‘spec’; try to resolve the first token to an
   existing delta; if found, ‘jpush’ its fully-numeric revno
   and return the "rest" of ‘spec’ (position after first token).
   If no delta can be determined, return NULL.  */
{
  const char delims[] = " \t\n:,;";
  char *eot, save;
  struct delta *cool;
  struct cbuf numrev;

  /* Locate bounds of token.  */
  SKIP_HWS (spec);
  eot = spec + strcspn (spec, delims);

  /* Save the delim, jam a terminating NUL there for the sake of
     ‘fully_numeric_no_k’, and restore it after the call.  Blech.  */
  save = *eot;
  *eot = '\0';
  cool = (fully_numeric_no_k (&numrev, spec)
          ? delta_from_ref (numrev.string)
          : NULL);
  *eot = save;

  if (! cool)
    return NULL;

  jpush (cool->num, js);
  SKIP_HWS (eot);
  return eot;
}

static char const *
getancestor (char const *r1, char const *r2)
/* Return the common ancestor of ‘r1’ and ‘r2’ if successful,
   ‘NULL’ otherwise.  Work reliably only if ‘r1’ and ‘r2’ are not
   branch numbers.   */
{
  char const *t1, *t2;
  int l1, l2, l3;
  char const *r;

  /* TODO: Don't bother saving in ‘PLEXUS’.  */

  l1 = countnumflds (r1);
  l2 = countnumflds (r2);
  if ((2 < l1 || 2 < l2) && !NUM_EQ (r1, r2))
    {
      /* Not on main trunk or identical.  */
      l3 = 0;
      while (NUMF_EQ (1 + l3, r1, r2)
             && NUMF_EQ (2 + l3, r1, r2))
        l3 += 2;
      /* This will terminate since ‘r1’ and ‘r2’ are not the
         same; see above.  */
      if (l3 == 0)
        {
          /* No common prefix; common ancestor on main trunk.  */
          t1 = TAKE (l1 > 2 ? 2 : l1, r1);
          t2 = TAKE (l2 > 2 ? 2 : l2, r2);
          r = NUM_LT (t1, t2) ? t1 : t2;
          if (!NUM_EQ (r, r1) && !NUM_EQ (r, r2))
            return str_save (r);
        }
      else if (!NUMF_EQ (1 + l3, r1, r2))
        return str_save (TAKE (l3, r1));
    }
  RERR ("common ancestor of %s and %s undefined", r1, r2);
  return NULL;
}

static bool
preparejoin (register char *argv, struct jstuff *js)
/* Parse join pairs from ‘argv’; ‘jpush’ their revision numbers.
   Set ‘js->lastidx’ to the last index of the list.
   Return ‘true’ if all went well, otherwise ‘false’.  */
{
  /* This is a two level parse, w/ whitespace complications:
     top-level is comma-delimited (straightforward), but
     join pairs are of the form: HWS REV1 HWS [":" HWS REV2 HWS]
     where HWS is optional horizontal whitespace (SPC, TAB).
     Also, the first "pair" can be comprised solely of REV1;
     in that case, the ":" and REV2 are also optional.

     Most of the HWS skipping is done in subroutine ‘addjoin’.

     TODO: Investigate why on addjoin "failure", it's OK to
     simply return ‘false’ (and not going through ‘done’).  */

  const char ks_comma[] = ",";
  char *s, *save, *j;
  bool rv = true;

  js->jstuff = make_space ("jstuff");
  js->head.next = NULL;
  js->tp = &js->head;
  if (! js->merge)
    {
      js->merge = ZLLOC (1, struct symdef);
      js->merge->meaningful = "merge";
    }

  js->lastidx = -1;
  for (s = argv;
       (j = strtok_r (s, ks_comma, &save));
       s = NULL)
    {
      if (!(j = addjoin (j, js)))
        return false;
      if (*j++ == ':')
        {
          SKIP_HWS (j);
          if (*j == '\0')
            goto incomplete;
          if (!(j = addjoin (j, js)))
            return false;
        }
      else
        {
          if (js->lastidx == 0)         /* first pair */
            {
              char const *two = js->tp->entry;

              /* Common ancestor missing.  */
              jpush (two, js);
              /* Derive common ancestor.  */
              if (! (js->tp->entry = getancestor (js->d->num, two)))
                rv = false;
            }
          else
            {
            incomplete:
              RFATAL ("join pair incomplete");
            }
        }
    }
  if (js->lastidx < 1)
    RFATAL ("empty join");

  js->ls = pointer_array (PLEXUS, 1 + js->lastidx);
  js->tp = js->head.next;
  for (int i = 0; i <= js->lastidx; i++, js->tp = js->tp->next)
    js->ls[i] = js->tp->entry;
  close_space (js->jstuff);
  js->jstuff = NULL;
  return rv;
}

/* Elements in the constructed command line prior to this index are
   boilerplate.  From this index on, things are data-dependent.  */
#define VX  3

static bool
buildjoin (char const *initialfile, struct jstuff *js)
/* Merge pairs of elements in ‘js->ls’ into ‘initialfile’.
   If ‘MANI (standard_output)’ is set, copy result to stdout.
   All unlinking of ‘initialfile’, ‘rev2’, and ‘rev3’
   should be done by ‘tempunlink’.  */
{
  char const *rev2, *rev3;
  int i;
  char const *cov[8 + VX], *mergev[11];
  char const **p;
  size_t len;
  char const *subs = NULL;

  rev2 = maketemp (0);
  rev3 = maketemp (3);      /* ‘buildrevision’ may use 1 and 2 */

  cov[1] = PEER_SUPER ();
  cov[2] = "co";
  /* ‘cov[VX]’ setup below.  */
  p = &cov[1 + VX];
  if (js->expand)
    *p++ = js->expand;
  if (js->suffix)
    *p++ = js->suffix;
  if (js->version)
    *p++ = js->version;
  if (js->zone)
    *p++ = js->zone;
  *p++ = quietarg;
  *p++ = REPO (filename);
  *p = '\0';

  mergev[1] = find_peer_prog (js->merge);
  mergev[2] = mergev[4] = "-L";
  /* Rest of ‘mergev’ setup below.  */

  i = 0;
  while (i < js->lastidx)
    {
#define ACCF(...)  accf (SINGLE, __VA_ARGS__)
      /* Prepare marker for merge.  */
      if (i == 0)
        subs = js->d->num;
      else
        {
          ACCF ("%s,%s:%s", subs, js->ls[i - 2], js->ls[i - 1]);
          subs = finish_string (SINGLE, &len);
        }
      diagnose ("revision %s", js->ls[i]);
      ACCF ("-p%s", js->ls[i]);
      cov[VX] = finish_string (SINGLE, &len);
      if (runv (-1, rev2, cov))
        goto badmerge;
      diagnose ("revision %s", js->ls[i + 1]);
      ACCF ("-p%s", js->ls[i + 1]);
      cov[VX] = finish_string (SINGLE, &len);
      if (runv (-1, rev3, cov))
        goto badmerge;
      diagnose ("merging...");
      mergev[3] = subs;
      mergev[5] = js->ls[i + 1];
      p = &mergev[6];
      if (BE (quiet))
        *p++ = quietarg;
      if (js->lastidx <= i + 2 && MANI (standard_output))
        *p++ = "-p";
      *p++ = initialfile;
      *p++ = rev2;
      *p++ = rev3;
      *p = '\0';
      if (DIFF_TROUBLE == runv (-1, NULL, mergev))
          goto badmerge;
      i = i + 2;
#undef ACCF
    }
  return true;

badmerge:
  FLOW (erroneous) = true;
  return false;
}

DECLARE_PROGRAM (co, BOG_FULL);

static int
co_main (const char *cmd, int argc, char **argv)
{
  int exitstatus = EXIT_SUCCESS;
  struct work work = { .force = false };
  struct jstuff jstuff;
  FILE *neworkptr = NULL;
  int lockflag = 0;                 /* -1: unlock, 0: do nothing, 1: lock.  */
  bool mtimeflag = false;
  char *a, *joinflag, **newargv;
  char const *author, *date, *rev, *state;
  char const *joinname, *newdate, *neworkname;
  /* 1 if a lock has been changed, -1 if error.  */
  int changelock;
  int expmode, r, workstatstat;
  bool tostdout, Ttimeflag;
  bool selfsame = false;
  char finaldate[DATESIZE];
#if OPEN_O_BINARY
  int stdout_mode = 0;
#endif
  struct wlink *deltas;                 /* Deltas to be generated.  */

  CHECK_HV (cmd);
  gnurcs_init (&program);
  memset (&jstuff, 0, sizeof (struct jstuff));

  setrid ();
  author = date = rev = state = NULL;
  joinflag = NULL;
  expmode = -1;
  tostdout = false;
  Ttimeflag = false;

  argc = getRCSINIT (argc, argv, &newargv);
  argv = newargv;
  while (a = *++argv, 0 < --argc && *a++ == '-')
    {
      switch (*a++)
        {

        case 'r':
        revno:
          chk_set_rev (&rev, a);
          break;

        case 'f':
          work.force = true;
          goto revno;

        case 'l':
          if (lockflag < 0)
            {
              PWARN ("-u overridden by -l.");
            }
          lockflag = 1;
          goto revno;

        case 'u':
          if (0 < lockflag)
            {
              PWARN ("-l overridden by -u.");
            }
          lockflag = -1;
          goto revno;

        case 'p':
          tostdout = true;
          goto revno;

        case 'I':
          BE (interactive) = true;
          goto revno;

        case 'q':
          BE (quiet) = true;
          goto revno;

        case 'd':
          if (date)
            redefined ('d');
          str2date (a, finaldate);
          date = finaldate;
          break;

        case 'j':
          if (*a)
            {
              if (joinflag)
                redefined ('j');
              joinflag = a;
            }
          break;

        case 'M':
          mtimeflag = true;
          goto revno;

        case 's':
          if (*a)
            {
              if (state)
                redefined ('s');
              state = a;
            }
          break;

        case 'S':
          selfsame = true;
          break;

        case 'T':
          if (*a)
            goto unknown;
          Ttimeflag = true;
          break;

        case 'w':
          if (author)
            redefined ('w');
          author = (*a)
            ? a
            : getcaller ();
          break;

        case 'x':
          jstuff.suffix = *argv;
          BE (pe) = a;
          break;

        case 'V':
          jstuff.version = *argv;
          setRCSversion (jstuff.version);
          break;

        case 'z':
          jstuff.zone = *argv;
          zone_set (a);
          break;

        case 'k':
          /* Set keyword expand mode.  */
          jstuff.expand = *argv;
          if (0 <= expmode)
            redefined ('k');
          if (0 <= (expmode = str2expmode (a)))
            break;
          /* fall into */
        default:
        unknown:
          bad_option (*argv);

        };
    }
  /* (End of option processing.)  */

  /* Now handle all filenames.  */
  if (FLOW (erroneous))
    cleanup (&exitstatus, &neworkptr);
  else if (argc < 1)
    PFATAL ("no input file");
  else
    for (; 0 < argc; cleanup (&exitstatus, &neworkptr), ++argv, --argc)
      {
        struct stat *repo_stat;
        char const *mani_filename;
        int kws;

        ffree ();

        if (pairnames
            (argc, argv,
             lockflag ? rcswriteopen : rcsreadopen,
             true, false) <= 0)
          continue;

        /* ‘REPO (filename)’ contains the name of the RCS file, and
           ‘FLOW (from)’ points at it.  ‘MANI (filename)’ contains the
           name of the working file.  Also, ‘REPO (stat)’ has been set.  */
        repo_stat = &REPO (stat);
        mani_filename = MANI (filename);
        kws = BE (kws);
        diagnose ("%s  -->  %s", REPO (filename),
                  tostdout ? "standard output" : mani_filename);

        workstatstat = -1;
        if (tostdout)
          {
#if OPEN_O_BINARY
            int newmode = kws == kwsub_b ? OPEN_O_BINARY : 0;
            if (stdout_mode != newmode)
              {
                stdout_mode = newmode;
                oflush ();
                setmode (STDOUT_FILENO, newmode);
              }
#endif
            neworkname = NULL;
            neworkptr = MANI (standard_output) = stdout;
          }
        else
          {
            workstatstat = stat (mani_filename, &work.st);
            if (!PROB (workstatstat) && SAME_INODE (REPO (stat), work.st))
              {
                RERR ("RCS file is the same as working file %s.",
                      mani_filename);
                continue;
              }
            neworkname = makedirtemp (true);
            if (!(neworkptr = fopen_safer (neworkname, FOPEN_W_WORK)))
              {
                if (errno == EACCES)
                  MERR ("permission denied on parent directory");
                else
                  syserror_errno (neworkname);
                continue;
              }
          }

        if (!REPO (tip))
          {
            /* No revisions; create empty file.  */
            diagnose ("no revisions present; generating empty revision 0.0");
            if (lockflag)
              PWARN ("no revisions, so nothing can be %slocked",
                     lockflag < 0 ? "un" : "");
            Ozclose (&FLOW (res));
            if (!PROB (workstatstat))
              if (!rmworkfile (&work))
                continue;
            changelock = 0;
            newdate = NULL;
          }
        else
          {
            struct cbuf numericrev;
            int locks = lockflag ? findlock (false, &jstuff.d) : 0;
            struct fro *from = FLOW (from);

            if (rev)
              {
                /* Expand symbolic revision number.  */
                if (!fully_numeric_no_k (&numericrev, rev))
                  continue;
              }
            else
              {
                switch (locks)
                  {
                  default:
                    continue;
                  case 0:
                    numericrev.string = GROK (branch) ? GROK (branch) : "";
                    break;
                  case 1:
                    numericrev.string = str_save (jstuff.d->num);
                    break;
                  }
              }
            /* Get numbers of deltas to be generated.  */
            if (! (jstuff.d = genrevs (numericrev.string, date, author,
                                          state, &deltas)))
              continue;
            /* Check reservations.  */
            changelock = lockflag < 0
              ? rmlock (jstuff.d)
              : (lockflag == 0
                 ? 0
                 : addlock_maybe (jstuff.d, selfsame, true));

            if (changelock < 0
                || (changelock && !checkaccesslist ())
                || PROB (dorewrite (lockflag, changelock)))
              continue;

            if (0 <= expmode)
              kws = BE (kws) = expmode;
            if (0 < lockflag && kws == kwsub_v)
              {
                RERR ("cannot combine -kv and -l");
                continue;
              }

            if (joinflag && !preparejoin (joinflag, &jstuff))
              continue;

            diagnose ("revision %s%s", jstuff.d->num,
                      0 < lockflag ? " (locked)" :
                      lockflag < 0 ? " (unlocked)" : "");
            SAME_AFTER (from, jstuff.d->text);

            /* Prepare to remove old working file if necessary.  */
            if (!PROB (workstatstat))
              if (!rmworkfile (&work))
                continue;

            /* Skip description (don't echo).  */
            write_desc_maybe (FLOW (to));

            BE (inclusive_of_Locker_in_Id_val) = 0 < lockflag;
            jstuff.d->name = namedrev (rev, jstuff.d);
            joinname = buildrevision (deltas, jstuff.d,
                                      joinflag && tostdout ? NULL : neworkptr,
                                      kws < MIN_UNEXPAND);
            if (FLOW (res) == neworkptr)
              FLOW (res) = NULL;             /* Don't close it twice.  */
            if (changelock && deltas->entry != jstuff.d)
              fro_trundling (true, from);

            if (PROB (donerewrite (changelock,
                                   file_mtime (Ttimeflag, repo_stat))))
              continue;

            if (changelock)
              {
                locks += lockflag;
                if (1 < locks)
                  RWARN ("You now have %d locks.", locks);
              }

            newdate = jstuff.d->date;
            if (joinflag)
              {
                newdate = NULL;
                if (!joinname)
                  {
                    aflush (neworkptr);
                    joinname = neworkname;
                  }
                if (kws == kwsub_b)
                  MERR ("merging binary files");
                if (!buildjoin (joinname, &jstuff))
                  continue;
              }
          }
        if (!tostdout)
          {
            mode_t m = WORKMODE (repo_stat->st_mode,
                                 !(kws == kwsub_v
                                   || (lockflag <= 0 && BE (strictly_locking))));
            time_t t = mtimeflag && newdate
              ? date2time (newdate)
              : TIME_UNSPECIFIED;
            aflush (neworkptr);
            IGNOREINTS ();
            r = chnamemod (&neworkptr, neworkname, mani_filename, 1, m,
                           make_timespec (t, ZERO_NANOSECONDS));
            keepdirtemp (neworkname);
            RESTOREINTS ();
            if (PROB (r))
              {
                syserror_errno (mani_filename);
                PERR ("see %s", neworkname);
                continue;
              }
            diagnose ("done");
          }
      }

  tempunlink ();
  Ozclose (&MANI (standard_output));
  gnurcs_goodbye ();
  return exitstatus;
}

static const uint8_t co_aka[13] =
{
  2 /* count */,
  2,'c','o',
  8,'c','h','e','c','k','o','u','t'
};

YET_ANOTHER_COMMAND (co);

/*:help
[options] file ...
Options:
  -f[REV]       Force overwrite of working file.
  -I[REV]       Interactive.
  -p[REV]       Write to stdout instead of the working file.
  -q[REV]       Quiet mode.
  -r[REV]       Normal checkout.
  -l[REV]       Like -r, but also lock.
  -u[REV]       Like -l, but unlock.
  -M[REV]       Reset working file mtime (relevant for -l, -u).
  -kSUBST       Use SUBST substitution, one of: kv, kvl, k, o, b, v.
  -dDATE        Select latest before or on DATE.
  -jJOINS       Merge using JOINS, a list of REV:REV pairs;
                this option is obsolete -- see rcsmerge(1).
  -sSTATE       Select matching state STATE.
  -S            Enable "self-same" mode.
  -T            Preserve the modification time on the RCS file
                even if it changes because a lock is added or removed.
  -wWHO         Select matching login WHO.
  -V            Obsolete; do not use.
  -VN           Emulate RCS version N.
  -xSUFF        Specify SUFF as a slash-separated list of suffixes
                used to identify RCS file names.
  -zZONE        Specify date output format in keyword-substitution
                and also the default timezone for -dDATE.

Multiple flags in {fIlMpqru} may be used, except for -r, -l, -u, which are
mutually exclusive.  If specified, REV can be symbolic, numeric, or mixed:
  symbolic -- must have been defined previously (see ci(1))
  $        -- determine the revision number from keyword values
              in the working file
  .N       -- prepend default branch => DEFBR.N
  BR.N     -- use this
  BR       -- latest revision on branch BR
If REV is omitted, take it to be the latest on the default branch.
*/

/* co.c ends here */
