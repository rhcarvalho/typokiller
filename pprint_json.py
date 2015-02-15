#!/usr/bin/env python
"""Quick and dirty command line helper to pretty-print JSON streams."""
import json, sys, collections
for line in sys.stdin:
    print json.dumps(json.loads(line, object_pairs_hook=collections.OrderedDict), indent=2)
