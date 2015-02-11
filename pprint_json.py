#!/usr/bin/env python
"""Quick and dirty command line helper to pretty-print JSON streams."""
import json, sys, collections
print json.dumps(json.load(sys.stdin, object_pairs_hook=collections.OrderedDict), indent=2)
