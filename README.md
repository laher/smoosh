# Smoosh

Smoosh is an experimental shell language written in Go. 

Smoosh is intended as primarily a programming language with shell features 'bolted on' (or _smooshed on top_).

_Smoosh is based on the sample 'monkey' language as defined in the book 'Writing an Interpreter in Go' by Thorsten Ball (Smoosh retains Monkey's MIT license)._

## A simple smoosh script

Smoosh will look a bit like this … (this isn't actually implemented yet)

```
   #!/usr/bin/smoosh
   var x = $("ls", "-1")
   echo x | $("grep", "1") \
          | >("out.log") \
          |2 $("tee", "err.log")
```

## Planned features

* Basic features similar to 'monkey', but `let` replaced with initialisation (`var` keyword) and plain reassignment.
* repl to be purposed as a shell environment - builtins such as `cd`, `alias`, `unalias`, `exit`
* `smoosh -fmt` to format a smoosh script to a standard
* Padding out several fundamental language features missing from monkey (floats, …)
* A rich standard library (based on parts of Go's standard lib)
* `$""` for running external commands. _NOTE: `$()` should be used for running commands inline, as in bash._
* `|` for piping. _Hopefully, typed pipes for slices._
* Unicode support.
* _Maybe_ unicode equivalents for readability. You'd type ascii as above and then `-fmt` would reformat to some equivalent like this ... maybe too crazy, eh
```
   echo x 🡒 $"grep", "1" ⤸
          ⤷ >"out.log" ⤸
          ⤷ₑ $"tee", "123"
```
* Go templating in place of bourne-style interpolation
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

