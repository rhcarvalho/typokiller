#!/usr/bin/env python
import json
import sys
from collections import OrderedDict

from enchant.checker import SpellChecker
from enchant.tokenize import EmailFilter, URLFilter


def spellcheck(pkg):
    """Return a list of documentation text with potential misspells."""
    chkr = SpellChecker("en_US", filters=[EmailFilter, URLFilter])
    misspelled_documentation = []
    for ident in pkg.get("Identifiers") or []:
        chkr.ignore_always(ident)
    for text in pkg.get("Documentation") or []:
        chkr.set_text(text.get("Content", ""))
        spelling_errors = [OrderedDict([
            ("Word", c.word),
            ("Offset", c.wordpos),
            ("Suggestions", c.suggest()),
        ]) for c in chkr]
        if spelling_errors:
            text["Misspellings"] = spelling_errors
            misspelled_documentation.append(text)
    return misspelled_documentation


if __name__ == "__main__":
    for line in sys.stdin:
        pkg = json.loads(line, object_pairs_hook=OrderedDict)
        misspelled_documentation = spellcheck(pkg)
        if misspelled_documentation:
            print json.dumps(OrderedDict([
                ("PackageName", pkg.get("PackageName", "")),
                ("Documentation", misspelled_documentation),
            ]), separators=(",", ":"))
