/* b-divvy.h --- dynamic memory manglement

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

#include <obstack.h>

struct divvy
{
  char const *name;
  struct obstack space;
  void *first;
  size_t count;
};

extern struct divvy *plexus;
extern struct divvy *single;

extern struct divvy *make_space (char const name[])
  ALL_NONNULL;
extern void *alloc (struct divvy *divvy, size_t len)
  ALL_NONNULL;
extern void *zlloc (struct divvy *divvy, size_t len)
  ALL_NONNULL;
extern char *intern (struct divvy *divvy, char const *s, size_t len)
  ALL_NONNULL;
extern void brush_off (struct divvy *divvy, void *ptr)
  ALL_NONNULL;
extern void forget (struct divvy *divvy)
  ALL_NONNULL;
extern void accf (struct divvy *divvy, char const *fmt, ...)
  ARG_NONNULL ((1, 2));
extern void accumulate_nbytes (struct divvy *divvy,
                               char const *start, size_t count)
  ALL_NONNULL;
extern void accumulate_byte (struct divvy *divvy, int c)
  ALL_NONNULL;
extern void accumulate_range (struct divvy *divvy,
                              char const *beg, char const *end)
  ALL_NONNULL;
extern void accs (struct divvy *divvy, char const *string)
  ALL_NONNULL;
extern char *finish_string (struct divvy *divvy, size_t *result_len)
  ALL_NONNULL;
extern void *pointer_array (struct divvy *divvy, size_t count)
  ALL_NONNULL;
extern void close_space (struct divvy *divvy)
  ALL_NONNULL;

/* Idioms.  */

#define PLEXUS  plexus
#define SINGLE  single

#define ZLLOC(n,type)          (zlloc (PLEXUS, sizeof (type) * n))
#define FALLOC(type)           (alloc (SINGLE, sizeof (type)))
#define FZLLOC(type)           (zlloc (SINGLE, sizeof (type)))

#define SHACCR(b,e)      accumulate_range (PLEXUS, b, e)
#define SHSTR(szp)       finish_string (PLEXUS, szp)
#define SHSNIP(szp,b,e)  (SHACCR (b, e), SHSTR (szp))

/* b-divvy.h ends here */
