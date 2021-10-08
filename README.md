# cuetils

CUE utilities and helpers for various tasks.

## Using

### As a library

Add to your project with `hof mod` or another method.

### As a tool

Clone the repository and add `$REPO/tools` to your `PATH`.
Make sure you have Python3 installed.

The tools are wrappers around the libraries.
Run their builtin help commands

__Structural:__

```
st -h
```

## Helpers

Many of these make use of the
["function" pattern](https://cuetorials.com/patterns/functions/).

### RecurseN

A function factory for bounded recursion, defaulting to 20.
This is a pattern to get around CUE's cycle detenction
by creating a struct with fields named for each iteration.
See https://cuetorials.com/deep-dives/recursion/ for more details.

```cue
#RecurseN: {
	#maxiter: uint | *20
	#funcFactory: {
		#next: _
		#func: _
	}

	for k, v in list.Range(0, #maxiter, 1) {
		#funcs: "\(k)": (#funcFactory & {#next: #funcs["\(k+1)"]}).#func
	}
	#funcs: "\(#maxiter)": null

	#funcs["0"]
}
```

The core of the bounded recursion is this structural comprehension (unrolled for loop).
The "recursive call" is made with the following pattern.

```cue
(#funcFactory & {#next: #funcs["\(k+1)"]}).#func
```

You can override the iterations with `{ #maxdepth: 100 }` at the point of
usage or by creating new helpers from the existing ones.
You may need to adjust this

- up for deep objects
- down if runtime is an issue

```cue
import "github.com/hofstadter-io/cuetils/structural"

#LeaguesDeep: structural.#Depth & { #maxdepth: 10000 }
```


### Structural

The following examples use

```cue
import (
	st "github.com/hofstadter-io/cuetils/structural"
)
```

__#Depth__ calculates the depth of an object

```cue
tree: {
	a: {
		foo: "bar"
		a: b: c: "d"
	}
	cow: "moo"
}

depth: (st.#Depth & { #in: tree }).out
depth: 5
```

__#Diff__ computes a semantic diff object

```cue
x: {
	a: "a"
	b: "b"
	d: "d"
	e: {
		a: "a"
		b: "b"
		d: "d"
	}
}

y: {
	b: "b"
	c: "c"
	d: "D"
	e:  {
		b: "b"
		c: "c"
		d: 1
	}
}

diff: (st.#Diff & { #X: x, #Y: y }).diff
diff: {
	"-": {
		a: "a"
		d: "d"
	}
	e: {
		"-": {
			a: "a"
			d: "d"
		}
		"+": {
			d: 1
			c: "c"
		}
	}
	"+": {
		d: "D"
		c: "c"
	}
}
```

__#Patch__ applies a diff object

```cue
x: {
	a: "a"
	b: "b"
	d: "d"
	e: {
		a: "a"
		b: "b"
		d: "d"
	}
}

p: {
	"-": {
		a: "a"
		d: "d"
	}
	e: {
		"-": {
			a: "a"
			d: "d"
		}
		"+": {
			d: 1
			c: "c"
		}
	}
	"+": {
		d: "D"
		c: "c"
	}
}

patch: (st.#Patch & { #X: x, #Y: y }).patch
patch: {
	b: "b"
	c: "c"
	d: "D"
	e:  {
		b: "b"
		c: "c"
		d: 1
	}
}
```

__#Pick__ extracts a subobject

```cue
```

__#Mask__ removes a subobject

```cue
```

### Custom Recursion

You can make new helpers by building on the `#RecurseN` pattern.
You need two definitions, a factory and the user facing, recursed version.

```cue
package structural

import (
	"list"

	"github.com/hofstadter-io/cuetils/recurse"
)

// A function factory
#depthF: {
	// always required
  #next: _
	
	// the actual computation, must be named #func
	#func: {
		// you can have any args
    #in: _
		// or internal helpers
		#basic: int|number|string|bytes|null
		
		// the result, can be named anything
    out: {
			if (#in & #basic) != _|_ { 1 }
			if (#in & #basic) == _|_ {
				list.Max([for k,v in #in {(#next & {#in: v}).out}]) + 1
			}
    }
  }
}

// The user facing, recursed version
#Depth: recurse.#RecurseN & {#funcFactory: #depthF}
```

The core of the recursive calling is:

```cue
(#next & {#in: v}).out
```

