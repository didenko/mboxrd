# Purpose

Split email archives downloaded from [Google Takeout (Download Your Data)][1] service into individual emails. Based on experimentation it looks like Google uses the _mboxrd_ dialect of _mbox_ format with CRLF lines as discussed at the [Wikipedia mbox article][2]

# License

The project is licensed under the [BSD 3-Clause License - see the `LICENSE.txt` file included with the package][3].

# Using the `mboxrd` package

The package provides both libraries and a buildable executable. See the code documentation on using the libraries.

[![GoDoc](https://godoc.org/github.com/didenko/mboxrd?status.svg)](https://godoc.org/github.com/didenko/mboxrd)

# Using the `mboxrd_split` executable

The executable takes the following parameters:

    -dir  <name>     : A directory to put the resulting messages to.
                       The directory must exist before running the program.

    -mbox <name>     : An mbox file to process and split into messages.

    -email <address> : An email which correspondence to be captured. Only
                       the actual address should be provided.

The program does not preserve unfinished last line of the last message in the archive. In the resulting files all message lines end with CRLF after the processing.

During the processing it created temporary message files and then moves them into the UTC-timestamped `.eml` file. If the destination filename is already taken by another message, then the later message does not override it. It is left in the temporary file and the error is printed to `stderr`.

Also a message stays in a temporary file if the program fails to construct a name for the message file. Some forwarded messages, for example, lack the `Date: ` header.

[1]: https://www.google.com/settings/takeout
[2]: https://en.wikipedia.org/wiki/Mbox#Family
[3]: ./LICENSE.txt

