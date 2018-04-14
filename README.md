# Smoosh

Smoosh is an experimental shell languae written in Go. 

Smoosh is intended a programming language with some esoteric shell features 'bolted on' (or _smooshed together_).

Smoosh is based on the sample 'monkey' language as described in the book 'Writing an Interpreter in Go'.

## A simple smoosh script

```
   #!/usr/bin/smoosh
   var x = $(ls -1)
   echo x | $"grep 1" \
          |> "out.log" \
          |2 $"tee err.log"
```

## Planned features

* Basic features similar to 'monkey', but `let` replaced with initialisation (`var` keyword) and plain reassignment.
* repl to be purposed as a shell environment - builtins such as `cd`, `alias`, `unalias`
* A rich standard library (based on parts of Go's standard lib)
* `smoosh -fmt` to format a smoosh script to a standard
* `$""` for running external commands. _NOTE: `$()` should be used for running commands inline, as in bash._
* `|` for piping.
* _Maybe_ unicode equivalents for readability. You type as above and then `-fmt` would reformat as follows ...
```
   echo x 🡒 $"grep 1" ⤸
          ⤷ "out.log" ⤸
          ⤷ₑ $"tee 123"
```

# Original Copyright from 'Writing An Interpreter In Go'

Thank you for purchasing "Writing An Interpreter In Go"!

...

Copyright © 2016-2017 Thorsten Ball
All rights reserved.
"Writing An Interpreter In Go" is copyright Thorsten Ball.

No part of this publication may be reproduced, stored in a retrieval system, or
transmitted, in any form, or by any means, electronic, mechanical, photocopying,
recording, or otherwise, without the prior consent of the publisher.

EXCEPT: the contents `code` folder are licensed under the MIT license
(https://opensource.org/licenses/MIT). See the `LICENSE` file in the `code`
folder.
