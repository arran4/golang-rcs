head	1.2;
access alice bob;
symbols
	beta:1.2.1.1 v1_0:1.1 v2_0:1.2;
locks
	alice:1.2;
strict;
comment	@# @;


1.2
date	2023.01.01.00.00.00;	author alice;	state Exp;
branches;
next	1.1;

1.1
date	2022.01.01.00.00.00;	author bob;	state Exp;
branches;
next	;


desc
@Access and Symbols@


1.2
log
@Rev 2@
text
@Content 2@


1.1
log
@Rev 1@
text
@Content 1@
