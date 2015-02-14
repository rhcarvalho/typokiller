typokiller
==========

Exterminate typos. Now.

Usage:

    go install ./...
    typokiller scan PATH/TO/GO/PKG | ./spellcheck.py | ./pprint_json.py | less
    typokiller scan PATH/TO/GO/PKG | ./spellcheck.py | typokiller fix

Or use the shortcut:

    ./killtypos PATH ...
