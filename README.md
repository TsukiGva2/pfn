# pfn

pfn is a transpiled programming language i made while trying to learn parsers, it's pretty basic and it isn't quite usable atm, it's still a work in progress.

it transpiles to python btw

[![Build Status](https://travis-ci.com/TsukiGva2/pfn.svg?branch=main)](https://travis-ci.com/TsukiGva2/pfn)

## why making yet another language?

i know, most of the stuff i do is mess around with
programming languages, but this is the last one, i
promise.

## About the language

### code sample?

here you go.

```py
.f (
  |x,y|

  -> (+ x y)
)

a:=2

(print (f a 5))
```

this transpiles to its equivalent python code

```py
def f(x,y):
  return (x+y)
  
a=2

(print((f(a,5))))
```

The output code isn't the most pretty or optimized, i plan to work on that part after i finish the language itself.

### Your code can be compact too

```py
.f ( |x,y| -> (+ x y) )
a := 2
(print (f a 5))
```

but try to not end up with code like this:

```py
.f(|x,y|->(+x y))a:=2(print(f a 5))
```

### why is the design mixed between lisp-like and whatever that other style is

I actualy spent 2 weeks trying to figure out how to parse expressions like 1 + 2 -2 + 2 / 3 * 2,
but i gave up after a few stack overflows, infinite loops and weird results.

I then switched to a new syntax, "+(x,y)", but after a few tests, i realized that it got pretty similar to the lisp-like syntax after some nesting, so i just switched to lisp-like.

## how to use this code?

make sure you have go installed, i do not plan to distribute precompiled binaries or something,
then clone the repo and run the code with

    $ go run .
    
this will start a repl where you can type some code, it is pretty simple,
type some code and when you're done, type "run", or, if you want to transpile single lines,
type "auto". When you're done, type "exit" or CTRL-C

you can also run code from a file with

    $ go run . filename

and if you want to just grab an executable, run

    $ go build .

and you will have an executable file named pc
