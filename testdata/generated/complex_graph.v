head	1.3;
access;
symbols;
locks;
comment	@# @;


1.3
date	2023.03.01.00.00.00;	author user;	state Exp;
branches;
next	1.2;

1.2
date	2023.02.01.00.00.00;	author user;	state Exp;
branches
	1.2.1.1
	1.2.2.1;
next	1.1;

1.1
date	2023.01.01.00.00.00;	author user;	state Exp;
branches;
next	;

1.2.1.1
date	2023.02.05.00.00.00;	author dev1;	state Exp;
branches;
next	1.2.1.2;

1.2.1.2
date	2023.02.06.00.00.00;	author dev1;	state Exp;
branches;
next	;

1.2.2.1
date	2023.02.07.00.00.00;	author dev2;	state Exp;
branches;
next	;


desc
@Complex Graph@


1.3
log
@Main 3@
text
@Main content 3@


1.2
log
@Main 2@
text
@Main content 2@


1.1
log
@Main 1@
text
@Main content 1@


1.2.1.1
log
@Branch 1.1@
text
@Branch content 1.1@


1.2.1.2
log
@Branch 1.2@
text
@Branch content 1.2@


1.2.2.1
log
@Branch 2.1@
text
@Branch content 2.1@
