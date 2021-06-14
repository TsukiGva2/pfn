let cLispFn  =	'\('
let cLispFn .=		'('
let cLispFn .=	'\)'
let cLispFn .=	'\@<='
let cLispFn .=	'[^()]'
let cLispFn .=	'\{-1,}'
let cLispFn .=	'\('
let cLispFn .=		'('
let cLispFn .=	'\|'
let cLispFn .=		'\s'
let cLispFn .=	'\|'
let cLispFn .=		')'
let cLispFn .=	'\)'
let cLispFn .=	'\@='

exe "syn match dLispFn display '" . cLispFn . "'"

hi def link dLispFn Function
