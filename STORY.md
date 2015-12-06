The Story...
============

The timeline:

- [Feb 2014](#before-typokiller-ever-existed)
- [Feb 2015](#the-new-tool)
- [Dec 2015](#redesign)


## Before `typokiller` Ever Existed

Even though for many times I've stumbled upon and corrected typos in open source
projects, `typokiller`'s story really started during a [Django Sprint in
Kraków](http://sprint.pykonik.org/), Poland, on the weekend of February 15-16th,
2014.

On that first morning, I paired with a just-met-friend and we decided to work on
Django's documentation together. For she was not a programmer, but a native
English speaker, we started by manually searching for and fixing common English
vs. British spelling cases.

We checked-in and submitted a few Pull Requests that later got merged into
Django:

- https://github.com/django/django/pull/2281
- https://github.com/django/django/pull/2282
- https://github.com/django/django/pull/2283
- https://github.com/django/django/pull/2288

Creating sequential PRs with small typo fixes wasn't good. After a couple of
hours of that and other attempts to find typos, I started to set myself for a
more ambitious goal. I wanted to kill all typos from the Django documentation.

To get there, and for not knowing of good existing tools, I hacked together a
Python script that would run all words from the documentation and source code by
a spell checker and create a SQLite database of possible typos. Later, I
interactively analyzed the database and used a text editor's search-replace
functionality to correct the misspellings.

There were too many false positives, which made me spend a considerable amount
of time to find and fix real typos. It was a tedious game of coming up with
heuristics to find what really needed fixing. Having to manually replace entries
was another time waster. I'd find a typo in the database, copy it, switch from
`sqlite3`'s shell to a text editor, open the search-replace dialog, paste, etc,
etc. until I got to another shell to review changes with `git add -p`, and,
finally, `git commit` the fixes.

Coincidentally, [@kwadrat](https://github.com/kwadrat) also took a "serious
approach" to killing typos that sprint, also with his own tools. Once we learned
that we had the same goal, we united forces to open PRs which killed more typos:

- https://github.com/django/django/pull/2384
- https://github.com/django/django/pull/2385
- https://github.com/django/django/pull/2386

That was it for 2014. I had no time to put on finding typos methodically,
however was still fixing typos here and there when I happened to spot them, not
using any particular automation tool.


## The New Tool

We're now exactly one year later, in February 2015. The idea of exterminating
typos efficiently was still in my head. One night I wrote the first commit to
the [project's repository](https://github.com/rhcarvalho/typokiller).
`typokiller` was born.

My goal is to have a handy tool to contribute to projects by efficiently
exterminating typos. Text processors have built-in spell checking for a long
time, there should be no excuses to having typos in text files.

It would be nice to check grammar too, that's a future goal.

The first time I talked about `typokiller` publicly was in a [Pykonik
meeting](http://blog.pykonik.org/2015/02/march-meeting-spotkanie-marcowe.html),
the group of Python users in Krákow.
[Here](http://www.slideshare.net/rhcarvalho/wsgi-andtypokiller) are the slides.

The fundamental design idea at that time was to create a pipeline flow. I wanted
to follow some of the UNIX philosophy and have small, single purpose
(sub)programs that could be composed among themselves and other UNIX utilities
by piping the output of one process into the input of another.

I also clearly needed some streamlined UI for going through potential typos and
fixing them without having to open a text editor.

Finally, improving on my original scripts, I knew I had to do some preprocessing
to eliminate a lot false positives.

After some, often sporadic, work on it, `typokiller` became usable. I even used
it to send more PRs to another open source project:

- https://github.com/openshift/origin/pull/1018
- https://github.com/openshift/openshift-docs/pull/160


## Redesign

I've redesigned `typokiller` in my head, and sometimes in a piece of paper, a
few times. In December 2015, I went ahead and started by re-documenting how
`typokiller` was to be used.

For one, I knew that I needed some form of data persistence. The pipeline idea
was fine, but the lack of persistence makes it really hard to work on larger
projects, and to be able to take back from where you left. No persistence also
means sadness if the program crashes before any changes were applied.

Reading about possible storage solutions, I decided to use
[Bolt](https://github.com/boltdb/bolt) for its simplicity and suitability for
the kind of data I needed to store.

I've introduced the concept of a "working project", to track files in multiple
locations in disk, and keep track of typo-hunting in them.

As with the first design, I still wanted a functional approach of data
transformation through a pipeline of actions and potentially long running worker
processes.

I wanted to get rid of Bash scripts and make it simpler to use `typokiller` as a
single all-in-one executable.

Learning from Computer-Aided Translation tools, I wanted something like
"translation memory": an ever growing database of typos and fixes.

And to improve on the number of false positives:

- As an alternative to spell checking all input words, try first to find known
  typos from a typo database;
- Sort potential typos by likelihood of being a real typo, so that real typos
  can be found and fixed even without going through all of the list of potential
  typos.

The redesign has followed a documentation-first approach. Before writing a
single line of code, I wrote the manual for how to use `typokiller` and
explained the main concepts.
