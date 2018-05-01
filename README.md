# Smoosh

Smoosh is an experimental shell language written in Go. 

Smoosh is intended as primarily a programming language with shell features 'bolted on' (or _smooshed on top_).

_Smoosh is based on the sample 'monkey' language as defined in the book 'Writing an Interpreter in Go' by Thorsten Ball (Smoosh retains Monkey's MIT license)._

## A simple smoosh script

Smoosh will look a bit like this … (this isn't actually implemented yet)

```
   var x = $("ls", "-1")
   echo x | $("grep", "1") \
          | >("out.log") \
          |2 $("tee", "err.log")
```

## Planned features

* Basic features similar to 'monkey'
* Repurpose monkey as a 'shell':
  - [X] accept smoosh piped in from STDIN
  - [X] take in a filename(s) for processing as a script
  - [ ] support/ignore a hashbang at the top of a file
  - [ ] support for piping external commands …
* Builtins:
  - [X] basic builtins such as `cd`, `exit`, `pwd`, `len`
  - [ ] `alias`, `unalias`
* Tooling:
  - [X] `smoosh-fmt` to format a smoosh script in a standard format
  - [X] Alternate REPL to print lexer results
  - [X] Alternate REPL to print AST as json
  - [ ] Line numbers (_a challenge for the reader_)
* Static types
  - [X] `let` replaced with initialisation (`var` keyword) and plain old reassignment
  - [ ] type checking
* Pad out some fundamental language features missing from monkey (floats, …)
  - [ ] floats/doubles
  - [ ] bitwise operators/logic?
  - [ ] bytes, reader, writer. Rune?
  - [ ] streams?
* A rich standard library (based on parts of Go's standard lib)
* Piping/execing primitives.
  AFAICT these primitives can be implemented as 'shorthands' or syntactic sugar for `os.Exec`
  - [ ] `$""` for running external commands. 
  - [ ] `$()` should be used for running commands inline, as in bash.
  - [ ] `|` for piping. _Hopefully, typed pipes for slices._
* Go templating in place of bourne-style interpolation
  - [ ] templating inside standard strings
  - [ ] multiline strings (syntax??)
* Unicode support.
  - [X] Parse smoosh in runes instead of bytes (_a challenge for the reader_)
  - [ ] _Maybe_ unicode equivalents for readability. You'd type ascii as above and then `-fmt` would reformat to some equivalent like this ... maybe too crazy, eh
```
   echo x 🡒 $"grep", "1" ⤸
          ⤷ >"out.log" ⤸
          ⤷ₑ $"tee", "123"
```
* Maybe remove parameter commas to increase shellishness

# Relevant excerpts from 'Writing An Interpreter In Go's README:

_Thank you for purchasing "Writing An Interpreter In Go"!_

… 

Copyright © 2016-2017 Thorsten Ball
All rights reserved.
"Writing An Interpreter In Go" is copyright Thorsten Ball.

… 

the contents `code` folder are licensed under the MIT license
(https://opensource.org/licenses/MIT). See the `LICENSE` file 

… 

