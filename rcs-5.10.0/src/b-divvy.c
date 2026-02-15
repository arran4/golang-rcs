/* b-divvy.c --- dynamic memory manglement

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
#include <stdarg.h>
#include <stdbool.h>
#include <stdlib.h>
#include "xalloc.h"
#include "b-complain.h"
#include "b-divvy.h"

struct divvy *plexus;
struct divvy *single;

#define TMALLOC(type)  xmalloc (sizeof (type))

#define obstack_chunk_alloc xmalloc
#define obstack_chunk_free free

struct divvy *
make_space (char const name[])
{
  struct divvy *divvy = TMALLOC (struct divvy);

  divvy->name = name;
  obstack_alloc_failed_handler = xalloc_die;
  obstack_init (&divvy->space);

  /* Set alignment to avoid segfault (on some hosts).
     The downside is wasted space on less-sensitive hosts.  */
  {
    size_t widest = (sizeof (void *) > sizeof (off_t))
      ? sizeof (void *)
      : sizeof (off_t);

    obstack_alignment_mask (&divvy->space) = widest - 1;
  }

  divvy->first = obstack_next_free (&divvy->space);
#ifdef DEBUG
  complain ("%s: %32s %p\n", name, "first", divvy->first);
#endif
  divvy->count = 0;
  return divvy;
}

void *
alloc (struct divvy *divvy, size_t len)
{
#ifdef DEBUG
  complain ("%s: %6u  what=???\n", divvy->name, len);
#endif
  divvy->count++;
  /* DWR: The returned memory is uninitialized.
     If you have doubts, use ‘zlloc’ instead.  */
  return obstack_alloc (&divvy->space, len);
}

void *
zlloc (struct divvy *divvy, size_t len)
{
  return memset (alloc (divvy, len), 0, len);
}

char *
intern (struct divvy *divvy, char const *s, size_t len)
{
#ifdef DEBUG
  complain ("%s: %6us %c%s%c\n", divvy->name, len,
            ('\0' == s[len]) ? '"' : '[',
            ('\0' == s[len]) ? s : "some bytes",
            ('\0' == s[len]) ? '"' : ']');
#endif
  divvy->count++;
  return obstack_copy0 (&divvy->space, s, len);
}

void
brush_off (struct divvy *divvy, void *ptr)
{
#ifdef DEBUG
  complain ("%s: %32s %p #%u\n", divvy->name, "brush-off", ptr, divvy->count);
#endif
  divvy->count--;
  obstack_free (&divvy->space, ptr);
}

void
forget (struct divvy *divvy)
{
#ifdef DEBUG
  complain ("%s: %32s %p (count=%u, room=%u)\n",
            divvy->name, "forget", divvy->first, divvy->count,
            obstack_room (&divvy->space));
#endif
  obstack_free (&divvy->space, divvy->first);
  divvy->count = 0;
}

void
accf (struct divvy *divvy, char const *fmt, ...)
{
  va_list args;

  va_start (args, fmt);
  obstack_vprintf (&divvy->space, fmt, args);
  va_end (args);
}

void
accumulate_nbytes (struct divvy *divvy, char const *start, size_t count)
{
  obstack_grow (&divvy->space, start, count);
}

void
accumulate_byte (struct divvy *divvy, int c)
{
  obstack_1grow (&divvy->space, c);
}

void
accumulate_range (struct divvy *divvy, char const *beg, char const *end)
{
  obstack_grow (&divvy->space, beg, end - beg);
}

void
accs (struct divvy *divvy, char const *string)
{
  obstack_grow (&divvy->space, string, strlen (string));
}

char *
finish_string (struct divvy *divvy, size_t *result_len)
{
  struct obstack *o = &divvy->space;
  char *rv;

  *result_len = obstack_object_size (o);
  obstack_1grow (o, '\0');
  rv = obstack_finish (o);
#ifdef DEBUG
  complain ("%s: %6ua \"%s\"\n", divvy->name, *result_len, rv);
#endif
  return rv;
}

void *
pointer_array (struct divvy *divvy, size_t count)
{
  struct obstack *o = &divvy->space;

#ifdef DEBUG
  complain ("%s: %6up (%u void*)\n", divvy->name,
            sizeof (void *) * count, count);
#endif
  while (count--)
    obstack_ptr_grow (o, NULL);
  return obstack_finish (o);
}

void
close_space (struct divvy *divvy)
{
  obstack_free (&divvy->space, NULL);
  divvy->count = 0;
  divvy->first = NULL;
  free (divvy);
}

/* b-divvy.c ends here */
