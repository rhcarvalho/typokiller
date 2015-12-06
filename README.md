typokiller
==========

Exterminate typos. Now.

`typokiller` is a tool for finding and fixing typos in text files, from source
code comments to documentation.


## Project's Background Story

Since you're here, you might be interested to know how this project was born. Go
on and read [the story behind `typokiller`](STORY.md).


## Hacking

Clone the repository:

```console
$ git clone https://github.com/rhcarvalho/typokiller.git
$ cd typokiller
```

Make sure you have installed:

* [Go](http://golang.org/)
* [Python](http://python.org/) (or [PyPy](http://pypy.org))

Tests showed that using PyPy gives a better spell checking throughput, and it
is, therefore, the recommended Python implementation.

Create a [virtualenv](https://virtualenv.pypa.io/en/latest/) (or [venv](https://docs.python.org/3/library/venv.html) if using Python 3.3+).

Activate the virtualenv and install the Python dependencies:

```console
$ pip install -r requirements.txt
```

Build the `typokiller` executable:

```console
$ ./build
```

Test it:

```console
$ ./test
```

## Usage

Quick start example:

```console
$ git clone https://github.com/@forkname@/@projectname@.git
$ cd @projectname@
$ typokiller init "My Project"
$ typokiller add .
$ typokiller console
...
$ git add .
$ git commit -m "Fix typos with typokiller"
$ git push
```

1. Fork an project on GitHub that you're interested in hunting down typos.
  - Replace `@forkname@` with your GitHub user name.
  - Replace `@projectname@` with the name of the project.
2. Initialize a `typokiller` project, optionally giving it a name.
3. Start the interactive console interface to find and fix typos.
4. Congratulations, you're done! Commit, push and submit a pull request.


### Pick a Target

`typokiller` was created to empower people to hunt and exterminate typos in open
source projects. Pick a project you'd like to contribute to, fork it, clone your
fork to you computer, and let's hunt down some typos!

Once you're done, submit a pull request to the chosen project.

I'd love to hear back from you if `typokiller` helped you exterminating typos,
which project did you contribute to, and what problems did you face.


### Create a New Project

The first step to using `typokiller` is to create a new project:

```console
$ typokiller init "My Project"
```

A new hidden directory and a project file will be created in the current
directory at `.typokiller/project`.

If a project name is not provided, one will be generated based on the current
directory name.

`typokiller` automatically stores in disk some information while you exterminate
typos. This is important to let you get back to where you were in a hunting
session without losing data or having to revisit typos in case a session was
interrupted.

Also, part of this information serves as a "memory" to help you quickly apply
fixes to typos you've handled before. This backs up features such as the [Typo
Database](#the-typo-database), the [Glossary](#the-glossary) and the [Auto
Replacement Table](#the-auto-replacement-table).

You can check your progress in the current project with the `status`
subcommand:

```console
$ typokiller status
Project: "My Project"
Locations: (empty)
Progress: 0% (fixed 0 out of 0 typos)

Use 'typokiller add' to add locations to this project.
```

Fair enough, we haven't really started just yet. Before we can start hunting
typos, we need to [add locations](#add-locations) to our project.

The project "progress" is measured as the number of actioned-upon typos over the
total number of typos. The total number of typos is given by the number of typos
pending [analysis](#analyzing-potential-typos) plus already [fixed
typos](#applying-changes). The total number of typos can be
[reset](#resetting-the-progress).


### Add Locations

A project may track one or more locations. A location is simply a path in the
file system. You add locations to a project using the `add` subcommand:

```console
$ typokiller add .
```

You can add files or directories. Adding a directory is equivalent to
recursively adding all files and subdirectories it contains.

If a location is in a Git repository, patterns ignored by Git will be
automatically ignored by `typokiller`. You can exclude additional patterns with
the `-x` option:

```console
$ typokiller add . -x boring-sub-directory
```

After adding a few locations, the project `status` might look like:

```console
$ typokiller status
Project: "My Project"
Locations:
  - "/home/user/project/part1": 2 subdirectories, 42 files (42 unread)
  - "/home/user/project/part2": 8 files (8 unread)
Progress: 0% (fixed 0 out of 0 typos)

No files were read yet, use 'typokiller read' to read the files.
Alternatively, use 'typokiller console' to hunt typos while files are read
automatically in the background.
```

Now, start `typokiller`'s interactive [console
interface](#the-console-interface) to find and fix typos.


### The Console Interface

The Console Interface is the headquarters of our typo killing activity. It is a
streamlined interface to present potential typos and act upon them.

Let's get it started:

```console
$ typokiller console
```

The Console will automatically start two new processes to [read
locations](#reading-locations) and [check segments](#checking-segments). This is
just a convenience and can be disabled with the `--no-read` and `--no-check`
options:

```console
$ typokiller console --no-read --no-check
```

### Analyzing Potential Typos

For each typo you analyze in the Console, you can take one of two kinds of
actions: replace or ignore.

Each kind of action can be applied to a single occurrence of a typo, to all
occurrences, or to all present and future occurrences. This allows you to
automatically resolve typos that haven't even been matched yet, by instantly
correcting common misspellings or ignoring certain terms.

For the sake of completeness, these are all the six possible actions:

- Replace occurrence of a typo with another string
- Replace all existing occurrences of a typo with another string
- Replace all existing and future occurrences of a typo with another string
  (add typo to the [Auto Replacement Table](#the-auto-replacement-table))
- Ignore occurrence of a typo
- Ignore all existing occurrences of a typo
- Ignore all existing and future occurrences of a typo
  (add typo as a new entry to the [Glossary](#the-glossary))

You can interrupt your analysis at any time. You're free to quit and come back
later. Meanwhile, the `status` subcommand helps you determine how the project
stands:

```console
$ typokiller status
Project: "My Project"
Locations:
  - "/home/user/project/part1": 2 subdirectories, 42 files
  - "/home/user/project/part2": 8 files
Progress: 42% (fixed 42 out of 100 typos)

No changes have been applied yet, use 'typokiller apply' to apply 5 pending
replacements in 2 files.
```


### Applying Changes

At any time you can apply the replace actions resulting from your analysis and
change the original files in disk. In other words, it may be done after you
finished analyzing all typos or at any earlier time.

Keep in mind that changing a file will automatically mark it as unread, so that
it will need to be re-read and re-checked. This is necessary to recompute
segment metadata after the file has changed.

You can apply changes directly from the [Console](#the-console-interface), or
from the command line:

```console
$ typokiller apply
```


### Advanced Usage

These are non-essential topics, for a more in-depth understanding of how
`typokiller` works and can be used.


#### Auto Apply

You may start a process for [applying changes](#applying-changes) and keep it
running indefinitely while applying changes as they are marked in the project:

```console
$ typokiller apply -w
```


#### Reading Locations

Before typo-hunting files added to a project, they need to be preprocessed to
collect some metadata. That's what "read" is all about.

Reading a location means opening a file, finding the interesting parts, and
splitting it into segments that can be later used by a
[_Checker_](#checking-segments).

Determining what the "interesting parts" are, depends on the _Reader_. Each
format supported by `typokiller` has a specific Reader. For example, the Go
Reader creates a segment for each comment block (representing portions of a Go
package's documentation), while ignoring function definitions and other parts of
the source code.

The Reader is determined by the location's format. See [Specifying Location
Format](#specifying-location-format) for details.

The Console can automatically start a process for reading locations. To read
locations from the command line:

```console
$ typokiller read
```

The `read` subcommand can optionally `add` more locations (and read them):

```console
$ typokiller read extra/path/1 extra/path/2
```

And you can [specify the format](#specifying-location-format) as in `add`:

```console
$ typokiller read --format=go extra/path/1 extra/path/2
```

Finally, you can start a long running process that will keep reading locations
as they are added (by another process):

```console
$ typokiller read -w
```


#### "Unreading" Locations

Locations are marked as unread when [applying changes](#applying-changes), so
that files can be re-checked for misspellings.

They can also be "unread" explicitly:

```console
$ typokiller unread part1 part2
```

To _unread_ all locations:

```console
$ typokiller unread --all
```


#### Specifying Location Format

Every location is associated with a certain format. By default, the format is
"Plain Text". To specify the format when [adding locations](#add-locations), use
the `--format` option:

```console
$ typokiller add --format=go extra/path/1 extra/path/2
```

Supported formats:

- Plain Text ("", "txt")
- Go documentation ("go")
- AsciiDoc ("adoc")

Planned support (send a Pull Request!):

- Markdown
- Python docstrings and comments
- Wikipedia dumps
- Diff/patch files


#### Checking Segments

There would be no typo extermination if it wasn't for something that finds
typos. That's the role of a _Checker_. A Checker is simply a strategy to detect
typos in text segments.

There are two strategies that can be used independently or, most often, together
to eliminate all those unwelcome typos:

- Spell Checker
- Typo Database

The Spell Checker uses the [enchant spell checking
library](https://pythonhosted.org/pyenchant/) to detect typos and suggest
corrections.

The [Typo Database](#the-typo-database) goes through a list of known typos and
try to find them in every segment. If a match is found, it will suggest
corrections from the database.

Normally, the Console runs in the background a process that uses concurrently
the two strategies above to detect typos. You can also start a process manually:

```console
$ typokiller check
```

To keep the process running indefinitely and checking new segments as the are
read, use the `-w` option:

```console
$ typokiller check -w
```

Finally, you can determine the strategy that will be used with the `--use`
option:

```console
$ typokiller check --use=typo-db
$ typokiller check --use=spell-checker
```

If `--use` is not provided, all strategies are used concurrently.


#### The Typo Database

Every project has a Typo Database. It is a table of known typos and a list of
spelling suggestions.

It can be used to quickly spot common typos while avoiding producing too many
false positives that a spell checker can bring.

Example:

Typo | Suggestions
---- | -----------
abandonned | abandoned, abandon
aberation | aberration, aeration
abilties | abilities
helo | hole, help, helot, hello, halo, hero, hell, held, helm

You can import and export a Typo Database using the CSV file format, where the
first column is a typo and the next columns are spelling suggestions.

In CSV, the example above would become:

```
abandonned,abandoned,abandon
aberation,aberration,aeration
abilties,abilities
helo,hole,help,helot,hello,halo,hero,hell,held,helm
```

To import known typos to the current project, use the `typo-db-import`
subcommand:

```console
$ typokiller typo-db-import ~/common-english-typos.csv ~/wikipedia-typos.csv
```

Conversely, you can export typos with `typo-db-export`:

```console
$ typokiller typo-db-export /path/to/typos.csv
```


#### The Glossary

Every project has a Glossary. Words in the glossary are automatically ignored by
[Checkers](#checking-segments).

You can import and export words using, respectively, the subcommands
`glossary-import` and `glossary-export`:

```console
$ typokiller glossary-import ~/project-glossary.txt
$ typokiller glossary-export ~/project-glossary.txt
```

The file format is plain text with one glossary entry per line.


#### The Auto Replacement Table

Every project has an Auto Replacement Table (ART). Words in the ART are
automatically replaced by their corresponding replacements.

Example:

Original | Replacement
-------- | -----------
colour | color
metre | meter
teh | the

You can import and export an ART using the CSV file format, where the first
column is a term to be replaced and the second column is the replacement.

In CSV, the example above would become:

```
colour,color
metre,meter
teh,thehelm
```

Import and export the ART using, respectively, the subcommands `art-import` and
`art-export`:

```console
$ typokiller art-import ~/replacement-table.csv
$ typokiller art-export ~/replacement-table.csv
```

#### Resetting the Progress

After you're done fixing typos, applying changes, committing and pushing, you
might want to reset your total number of typos.

You'd normally do that after pulling in a new version of your files from version
control, so that the progress reported by `typokiller status` becomes more
helpful.

Example:

```console
$ typokiller status
Project: "My Project"
Locations:
  - "/home/user/project/part1": 2 subdirectories, 42 files
  - "/home/user/project/part2": 8 files
Progress: 100% (fixed 100 out of 100 typos)

You're awesome! All typos have been exterminated.
$ git pull
$ typokiller reset
$ typokiller status
Project: "My Project"
Locations:
  - "/home/user/project/part1": 2 subdirectories, 42 files (42 unread)
  - "/home/user/project/part2": 8 files (8 unread)
Progress: 0% (fixed 0 out of 0 typos)

No files were read yet, use 'typokiller read' to read the files.
Alternatively, use 'typokiller console' to hunt typos while files are read
automatically in the background.
```

The `reset` subcommand marks all locations as unread, and reset the typo counter
to 0. All the typo-hunting memory is left intact: Typo Database, Glossary, Auto
Replacement Table.
