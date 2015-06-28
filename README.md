typokiller
==========

Exterminate typos. Now.


Hacking
-------

Clone the repository:

```bash
$ git clone https://github.com/rhcarvalho/typokiller.git
```

Build and runtime requirements:

* Proper [Go](http://golang.org/doc/install) [environment](http://golang.org/doc/code.html#GOPATH) (tested with Go 1.4.1, probably works with older versions)
* [Python](http://python.org/) 2.x (seems to work faster with [PyPy](http://pypy.org))

It's recommended to install PyPy and use it inside a [virtualenv](https://virtualenv.pypa.io/en/latest/).

Activate the virtualenv and run:

```bash
$ pip install -r requirements.txt
```

Build and install Go executable:

```bash
$ cd typokiller
$ go get ./...
```


Usage
-----

The shortcut:

```bash
$ ./killtypos PATH ...
```

This will build and install `typokiller`, read the documentation of the Go packages in PATH(s), spellcheck it all, and present a terminal-based UI for fixing typos.

Integrate with `find` for great profit:

```bash
$ ./killtypos $(find /PATH/TO/GIT/REPO -type d -not \( -name .git -prune -o -name Godeps -prune \))
```

This will find typos in all directories under `/PATH/TO/GIT/REPO` (inclusive), ignoring anything under `.git` and `Godeps`.

You can also use the parts separately for debugging or integration with other UNIX tools:

```bash
# normal usage:
$ typokiller read /PATH/TO/GO/PKG | ./spellcheck.py | typokiller fix

# inspect spellcheck results manually:
$ typokiller read /PATH/TO/GO/PKG | ./spellcheck.py | ./pprint_json.py | less

# limit number of packages:
$ typokiller read /PATH/TO/GO/PKG | head -n 20 | ./spellcheck.py | ./pprint_json.py | less
```
