/* b-yacmd.h --- yet another command

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

typedef int (submain_t) (const char *cmd, int argc, char **argv);

struct yacmd
{
  submain_t *func;
  const uint8_t *aka;
  struct program *pr;
};

#define YA(prog)  ya_ ## prog

#define YET_ANOTHER_COMMAND(prog)               \
  const struct yacmd YA (prog) =                \
  {                                             \
    .func = prog ## _main,                      \
    .aka = prog ## _aka,                        \
    .pr = &program                              \
  }

/* b-yacmd.h ends here */
