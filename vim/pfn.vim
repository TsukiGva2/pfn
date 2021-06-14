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

syn match dFunction "\(\.\)\@<=[a-zA-Z0-9]\+\(\s\{0,}(\)\@="
syn match dOperator "->\|[.|]\|-\(\p\)\@=\|:="
syn region dComment start="#" end="$"
syn region dString start='"' end='"'
syn keyword dKw when do end

" Integer with - + or nothing in front
syn match dNumber '\d\+'
syn match dNumber '[-+]\d\+'

" Floating point number with decimal no E or e 
syn match dNumber '[-+]\d\+\.\d*'

" Floating point like number with E and no decimal point (+,-)
syn match dNumber '[-+]\=\d[[:digit:]]*[eE][\-+]\=\d\+'
syn match dNumber '\d[[:digit:]]*[eE][\-+]\=\d\+'

" Floating point like number with E and decimal point (+,-)
syn match dNumber '[-+]\=\d[[:digit:]]*\.\d*[eE][\-+]\=\d\+'
syn match dNumber '\d[[:digit:]]*\.\d*[eE][\-+]\=\d\+'

hi def link dLispFn   Function
hi def link dFunction Function
hi def link dComment  Comment
hi def link dKw       Keyword
hi def link dOperator Operator
hi def link dString   String
hi def link dNumber   Number

