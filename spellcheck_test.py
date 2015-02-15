#!/usr/bin/env python
import unittest
import json

from spellcheck import spellcheck


class TestSpellcheck(unittest.TestCase):
    def setUp(self):
        with open("testdata/read.json") as f:
            self.pkgs = [json.loads(line) for line in f]

    def runTest(self):
        words = [
            [misspelling["Word"] for misspelling in text["Misspellings"]]
            for pkg in self.pkgs
            for text in spellcheck(pkg)
        ]
        self.assertItemsEqual(words, [
            ["interfeice"],
            ["typokiller"],
            ["helo", "Gophr"],
            ["typokiller"],
        ])


if __name__ == "__main__":
    unittest.main()
