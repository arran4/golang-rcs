/* b-fb.h --- basic file operations

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

extern int change_mode (int fd, mode_t mode);
exiting
extern void Ierror (void);
extern void testIerror (FILE *f)
  ALL_NONNULL;
exiting
extern void Oerror (void);
extern void testOerror (FILE *o)
  ALL_NONNULL;
extern FILE *fopen_safer (char const *filename, char const *type)
  ALL_NONNULL;
extern void Ozclose (FILE **p)
  ALL_NONNULL;
extern void aflush (FILE *f)
  ALL_NONNULL;
extern void oflush (void);
extern void afputc (int c, FILE *f)
  ALL_NONNULL;
extern void newline (FILE *f)
  ALL_NONNULL;
extern void aputs (char const *s, FILE *iop)
  ALL_NONNULL;
extern void aprintf (FILE *iop, char const *fmt, ...)
  ARG_NONNULL ((1, 2))
  printf_string (2, 3);

/* b-fb.h ends here */
