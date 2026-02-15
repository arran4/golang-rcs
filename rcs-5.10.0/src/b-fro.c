/* b-fro.c --- read-only file

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
#include <stdlib.h>
#include <errno.h>
#include <string.h>                     /* strchr */
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/stat.h>
#ifdef HAVE_SYS_MMAN_H
#include <sys/mman.h>
#endif
#include <unistd.h>
#include "unistd-safer.h"
#include "b-complain.h"
#include "b-divvy.h"
#include "b-fb.h"
#include "b-fro.h"
#include "b-isr.h"

static void
mmap_deallocate (struct fro *f)
{
  if (PROB (munmap (f->base, f->lim - f->base)))
    fatal_sys ("munmap");
}

struct fro *
fro_open (char const *name, char const *type, struct stat *status)
/* Open ‘name’ for reading, return its descriptor, and set ‘*status’.  */
{
  struct fro *f;
  FILE *stream;
  struct stat st;
  off_t s;
  int fd = fd_safer (open (name, O_RDONLY
                           | (OPEN_O_BINARY && strchr (type, 'b')
                              ? OPEN_O_BINARY
                              : 0)));
  bool unlimited = MEMORY_UNLIMITED == BE (mem_limit);

  if (PROB (fd))
    return NULL;
  if (!status)
    status = &st;
  if (PROB (fstat (fd, status)))
    fatal_sys (name);
  if (!S_ISREG (status->st_mode))
    {
      PERR ("`%s' is not a regular file", name);
      close (fd);
      errno = EINVAL;
      return NULL;
    }
  s = status->st_size;

  f = FZLLOC (struct fro);
  f->end = s;

  /* Determine the read method.  */
  f->rm = (unlimited
           || (status->st_size >> 10) < BE (mem_limit))
    ? (MMAP_SIGNAL && status->st_size
       ? RM_MMAP
       : RM_MEM)
    : RM_STDIO;

#define STUBBORNLY_RETRY_MAYBE(METHOD)  do      \
    {                                           \
      if (unlimited)                            \
        {                                       \
          f->rm = METHOD;                       \
          goto retry;                           \
        }                                       \
      fatal_sys (name);                         \
    }                                           \
  while (0)

 retry:
  switch (f->rm)
    {
    case RM_MMAP:
      if (MMAP_SIGNAL)
        {
          if (s != status->st_size)
            PFATAL ("%s: too large", name);
          f->stream = NULL;
          f->deallocate = NULL;
          ISR_DO (CATCHMMAPINTS);
          f->base = mmap (NULL, s, PROT_READ, MAP_SHARED, fd, 0);
          if (f->base == MAP_FAILED)
            STUBBORNLY_RETRY_MAYBE (RM_MEM);
          /* On many hosts, the superuser can mmap an NFS file
             it can't read.  So access the first page now, and
             print a nice message if a bus error occurs.  */
          if (has_NFS)
            access_page (ISR_SCRATCH, name, f->base);
          f->deallocate = mmap_deallocate;
          f->ptr = f->base;
          f->lim = f->base + s;
          fro_trundling (true, f);
          break;
        }
      else
        {
          /* fall through */
          ;
        }

    case RM_MEM:
      /* Read it into main memory all at once; this is
         the simplest substitute for memory mapping.  */
      if (!s)
        /* We used to set this to some non-‘NULL’ value.
           What's wrong with using ‘NULL’?  */
        f->base = NULL;
      else
        {
          ssize_t r;
          char *bufptr;
          size_t bufsiz = s;

          f->base = alloc (SINGLE, s);
          bufptr = f->base;
          do
            {
              if (PROB (r = read (fd, bufptr, bufsiz)))
                STUBBORNLY_RETRY_MAYBE (RM_STDIO);

              if (!r)
                {
                  /* The file must have shrunk!  */
                  status->st_size = s -= bufsiz;
                  bufsiz = 0;
                }
              else
                {
                  bufptr += r;
                  bufsiz -= r;
                }
            }
          while (bufsiz);
          if (PROB (lseek (fd, 0, SEEK_SET)))
            STUBBORNLY_RETRY_MAYBE (RM_STDIO);
        }
      f->ptr = f->base;
      f->lim = f->base + s;
      break;

    case RM_STDIO:
      if (!(stream = fdopen (fd, type)))
        fatal_sys (name);
      f->stream = stream;
      break;
    }

  f->fd = fd;
  return f;

#undef STUBBORNLY_RETRY_MAYBE
}

void
fro_close (struct fro *f)
{
  int res = -1;

  if (!f)
    return;
  switch (f->rm)
    {
    case RM_MMAP:
    case RM_MEM:
      if (f->deallocate)
        (*f->deallocate) (f);
      f->base = NULL;
      res = close (f->fd);
      break;
    case RM_STDIO:
      res = fclose (f->stream);
      break;
    }
  /* Don't use ‘PROB (res)’ here; ‘fclose’ may return EOF.  */
  if (res)
    Ierror ();
  f->fd = -1;
}

void
fro_zclose (struct fro **p)
{
  fro_close (*p);
  *p = NULL;
}

off_t
fro_tello (struct fro *f)
{
  off_t rv = 0;

  switch (f->rm)
    {
    case RM_MMAP:
    case RM_MEM:
      rv = f->ptr - f->base;
      break;
    case RM_STDIO:
      rv = ftello (f->stream);
      break;
    }
  return rv;
}

void
fro_move (struct fro *f, off_t change)
/* If ‘change’ is less than 0, seek relative to the current position.
   Otherwise, seek to the absolute position.  */
{
  switch (f->rm)
    {
    case RM_MMAP:
    case RM_MEM:
      f->ptr = change + (0 > change
                         ? f->ptr
                         : f->base);
      break;
    case RM_STDIO:
      if (PROB (fseeko (f->stream, change, 0 > change ? SEEK_CUR : SEEK_SET)))
        Ierror ();
      break;
    }
}

#define GETBYTE_BODY()                          \
  switch (f->rm)                                \
    {                                           \
    case RM_MMAP:                               \
    case RM_MEM:                                \
      if (f->ptr == f->lim)                     \
        DONE ();                                \
      *c = *f->ptr++;                           \
      break;                                    \
    case RM_STDIO:                              \
      {                                         \
        FILE *stream = f->stream;               \
        int maybe = getc (stream);              \
                                                \
        if (EOF == maybe)                       \
          {                                     \
            testIerror (stream);                \
            DONE ();                            \
          }                                     \
        *c = maybe;                             \
      }                                         \
      break;                                    \
    }

bool
fro_try_getbyte (int *c, struct fro *f)
/* Try to get another byte from ‘f’ and set ‘*c’ to it.
   If at EOF, return true.  */
{
#define DONE()  return true
  GETBYTE_BODY ();
#undef DONE
  return false;
}

void
fro_must_getbyte (int *c, struct fro *f)
/* Try to get another byte from ‘f’ and set ‘*c’ to it.
   If at EOF, signal "unexpected end of file".  */
{
#define DONE()  SYNTAX_ERROR ("unexpected end of file")
  GETBYTE_BODY ();
#undef DONE
}

#ifdef HAVE_MADVISE
#define USED_IF_HAVE_MADVISE
#else
#define USED_IF_HAVE_MADVISE  RCS_UNUSED
#endif

void
fro_trundling (bool sequential USED_IF_HAVE_MADVISE, struct fro *f)
/* Advise the mmap machinery (if applicable) that access to ‘f’
   is sequential if ‘sequential’, otherwise normal.  */
{
  switch (f->rm)
    {
    case RM_MMAP:
#ifdef HAVE_MADVISE
      madvise (f->base, f->lim - f->base,
               sequential ? MADV_SEQUENTIAL : MADV_NORMAL);
#endif
      break;
    case RM_MEM:
    case RM_STDIO:
      break;
    }
}

void
fro_spew_partial (FILE *to, struct fro *f, struct range *r)
{
  switch (f->rm)
    {
    case RM_MMAP:
    case RM_MEM:
      /* TODO: Handle range larger than ‘size_t’.  */
      awrite (f->base + r->beg, r->end - r->beg, to);
      if (f->end == r->end)
        f->ptr = f->lim;
      break;
    case RM_STDIO:
      {
#define MEMBUFSIZ  (8 * BUFSIZ)
        char buf[MEMBUFSIZ];
        size_t count;
        off_t pos = r->beg;

        fseeko (f->stream, pos, SEEK_SET);
        while (pos < r->end)
          {
            if (!(count = fread (buf, sizeof (*buf),
                                 (pos < r->end - MEMBUFSIZ
                                  ? MEMBUFSIZ
                                  : r->end - pos),
                                 f->stream)))
              {
                testIerror (f->stream);
                return;
              }
            awrite (buf, count, to);
            pos += count;
          }
#undef MEMBUFSIZ
      }
      break;
    }
}

void
fro_spew (struct fro *f, FILE *to)
/* Copy the remainder of file ‘f’ to ‘to’.  */
{
  struct range finish =
    {
      .beg = f->verbatim,
      .end = f->end
    };

  fro_spew_partial (to, f, &finish);
  f->verbatim = f->end;
}

struct cbuf
string_from_atat (struct divvy *space, struct atat const *atat)
{
  struct fro *f = atat->from;
  size_t count = atat->count;
  struct cbuf cb;
  size_t i;

#define RBEG(i)  (1 + (i ? atat->holes[i - 1] : atat->beg))

  switch (f->rm)
    {
    case RM_MMAP:
    case RM_MEM:
      for (i = 0; i < count; i++)
        {
          off_t rbeg = RBEG (i);
          char const *beg = f->base + rbeg;
          off_t len = atat->holes[i] - rbeg;

          while (SSIZE_MAX < len)
            {
              accumulate_nbytes (space, beg, SSIZE_MAX);
              len -= SSIZE_MAX;
              beg += SSIZE_MAX;
            }
          accumulate_nbytes (space, beg, len);
        }
      break;
    case RM_STDIO:
      {
        FILE *stream = f->stream;
        off_t was = ftello (stream);

        for (i = 0; i < count; i++)
          {
            off_t pos = RBEG (i);

            fseeko (stream, pos, SEEK_SET);
            while (pos++ < atat->holes[i])
              accumulate_byte (space, getc (f->stream));
          }
        fseeko (stream, was, SEEK_SET);
      }
      break;
    }
  cb.string = finish_string (space, &cb.size);
  return cb;

#undef RBEG
}

void
atat_put (FILE *to, struct atat const *atat)
{
  struct range range =
    {
      .beg = atat->beg,
      .end = ATAT_TEXT_END (atat)
    };

  fro_spew_partial (to, atat->from, &range);
}

void
atat_display (FILE *to, struct atat const *atat, bool ensure_end_nl)
{
  for (size_t i = 0; i < atat->count; i++)
    {
      struct range range =
        {
          .beg = 1 + (i ? atat->holes[i - 1] : atat->beg),
          .end = atat->holes[i]
        };

      fro_spew_partial (to, atat->from, &range);
    }

  /* Don't bother with trailing '\n' output if not requested,
     or if the atat is empty.  */
  if (! ensure_end_nl
      || (1 == atat->count
          && atat->beg + 1 == atat->holes[0]))
    return;

  {
    struct fro *f = atat->from;
    off_t pos = atat->holes[atat->count - 1] - 1;
    char lc = '\0';

    switch (f->rm)
      {
      case RM_MMAP:
      case RM_MEM:
        lc = f->base[pos];
        break;
      case RM_STDIO:
        {
          FILE *stream = f->stream;
          off_t was = ftello (stream);

          fseeko (stream, pos, SEEK_SET);
          lc = fgetc (stream);
          fseeko (stream, was, SEEK_SET);
        }
        break;
      }

    if ('\n' != lc)
      newline (to);
  }
}

/* b-fro.c ends here */
