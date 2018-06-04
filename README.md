# Smoosh

Smoosh is an experimental shell language written in Go. 

Smoosh is intended as primarily a programming language with shell-like features 'bolted on' (or _smooshed on top_).

_Smoosh is based on the sample 'monkey' language as defined in the book 'Writing an Interpreter in Go' by Thorsten Ball (Smoosh retains Monkey's MIT license)._

_NOTE: Smoosh is cross-plaform. For portability reasons, smoosh doesn't actually use linux pipes. It uses the io.Readers/io.Writers supplied by `exec.Cmd`, and pipes data through these. Performance and memory consumption seem acceptable so far._

## A simple smoosh script

Smoosh will look a bit like this … (this isn't completely implemented yet)

```
   var x = $(`ls -1`)
   echo x | $(`grep 1`) | w("out.log", "err.log")
```

## Planned features

* Basic features similar to 'monkey'
* Repurpose monkey as a 'shell':
  - [X] accept smoosh piped in from STDIN
  - [X] take in a filename(s) for processing as a script
  - [X] support/ignore a hashbang at the top of a file
  - [X] support for piping external commands …
  - [X] Backticks for succinctness
  - [ ] support for exit codes, signals
  - [ ] support for interactive commands
  - [ ] history
  - [ ] key mappings for repl (up-arrow, home, end, etc)
  - [ ] pimping: shell completion, colours, etc
* Builtins:
  - [X] basic builtins such as `cd`, `exit`, `pwd`, `len`
  - [X] Redirection helpers which avoid `>`/`<` symbols (avoid gt/lt collisions). i.e. `r()` and `w()`
  - [X] `w(a, "x.txt")` for append-to-file
  - [X] 'coreutils' (roughly)
    - [X] basename, dirname
    - [X] cat
    - [X] cp, mv, rm
    - [X] grep
    - [X] gunzip, gzip
    - [X] zip, unzip
    - [X] head, tail
    - [X] ls
    - [X] sleep
    - [X] tee
    - [X] touch
    - [X] wc
    - [X] which
  - [ ] `alias`, `unalias`
  - [ ] pipe stuff e.g. `red(2,1)` for redirection
  - [ ] process-handling stuff (signals, exit codes, async processing ...)
  - [ ] file-handling stuff (exists, is-directory, r/w/x permissions)
  - [ ] env stuff
* Tooling:
  - [X] `smoosh -fmt` to format a smoosh script in a standard format
  - [X] Alternate REPL to print lexer results
  - [X] Alternate REPL to print AST as json
  - [X] Line numbers (_a challenge for the reader_)
* Static types
  - [X] `let` replaced with initialisation (`var` keyword) and plain old reassignment
  - [X] type checking
* Pad out some fundamental language features missing from monkey (floats, …)
  - [ ] floats/doubles
  - [X] loops
  - [X] comments
  - [ ] bitwise operators/logic?
  - [ ] bytes, reader, writer. Rune? streams?
* A standard library (based on parts of Go's standard lib)
  - [X] A single example (http.Get)
  - [ ] Some kind of hook into Go's stdlib (without wrapping every dam thing)
* Dependencies
  - [ ] including files/packages
  - [ ] referencing 3rd party Go packages - are plugins needed here?
* Piping/execing primitives.
  AFAICT these primitives can be implemented as 'shorthands' or syntactic sugar for `os.Exec`
  - [X] `$("")` for running external commands. 
  - [X] `|` for piping. _Hopefully, typed pipes for slices._
* Go templating in place of bourne-style interpolation
  - [X] templating inside standard strings
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

