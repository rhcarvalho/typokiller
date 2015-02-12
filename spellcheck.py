#!/usr/bin/env python
import json
import sys
from collections import OrderedDict

from enchant.checker import SpellChecker
from enchant.tokenize import EmailFilter, URLFilter


def spellcheck_packages(pkgs):
    misspelled_packages = []
    for pkg in pkgs:
        misspelled_comments = spellcheck(pkg)
        if misspelled_comments:
            misspelled_packages.append(OrderedDict([
                ("PackageName", pkg.get("PackageName", "")),
                ("Comments", misspelled_comments),
            ]))
    return misspelled_packages

def spellcheck(pkg):
    chkr = SpellChecker("en_US", filters=[EmailFilter, URLFilter])
    misspelled_comments = []
    for ident in pkg.get("Identifiers") or []:
        chkr.ignore_always(ident)
    for comment in pkg.get("Comments") or []:
        chkr.set_text(comment.get("Text", ""))
        spelling_errors = [OrderedDict([
            ("Word", c.word),
            ("Offset", c.wordpos),
            ("Suggestions", c.suggest()),
        ]) for c in chkr]
        if spelling_errors:
            comment["SpellingErrors"] = spelling_errors
            misspelled_comments.append(comment)
    return misspelled_comments


if __name__ == "__main__":
    pkgs = json.load(sys.stdin, object_pairs_hook=OrderedDict)
    misspelled_packages = spellcheck_packages(pkgs)
    print json.dumps(misspelled_packages, separators=(",", ":"))
