sublimegen
---------------

This is a tool that will convert a BNF (conforming to the GOCC syntax) for a language to a
SublimeText Syntax Highlighting Plugin. It uses gocc (https://code.google.com/p/gocc/), a parser generator written in Go.


dependecies
---------------

- golang-pkg-pcre (https://github.com/glenn-brown/golang-pkg-pcre)
        - libpcre++-dev
- gouuid (for License see gouuid_LICENSE, https://github.com/nu7hatch/gouuid)
- greenery (for License see greenery_LICENSE, original: https://github.com/ferno/greenery)
        - please note that the `greenery` that is being used is not the original. the
        custom `greenery` can be found under `github.com/AdeebNqo/sublimegen/src/greenery`.
- >=Python 2.7.6


additional
---------------

The python scripts (test_scripts/generate-*.py) are used for testing. The tool does not depend on them.
The scripts depend on xeger (which can be found in rstr, https://pypi.python.org/pypi/rstr/2.1.2).
