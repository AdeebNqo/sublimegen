sublimegen
---------------

[![Build Status](https://drone.io/github.com/AdeebNqo/sublimegen/status.png)](https://drone.io/github.com/AdeebNqo/sublimegen/latest)


This is a tool that will convert a BNF (conforming to the GOCC syntax) for a language to a
SublimeText Syntax Highlighting Plugin. It uses [gocc](https://code.google.com/p/gocc/), a parser generator written in Go.

[![forthebadge](http://forthebadge.com/images/badges/gluten-free.svg)](http://forthebadge.com)
[![forthebadge](http://forthebadge.com/images/badges/does-not-contain-msg.svg)](http://forthebadge.com)

How to use
-----------
If we assume that we generated a binary file called `sublimegen`. Running it for the language
`Calc` whose configuration files can be found in `languagefiles/calc_languagefiles` is ass follows:

```./sublimegen -fileTypes "calc" -name "Calc" -scopeName "source.calc" -source languagefiles/calc_languagefiles/calc.bnf -scopes languagefiles/calc_languagefiles/scopes.json -orderregex 1```

Dependencies
---------------

- [golang-pkg-pcre](https://github.com/glenn-brown/golang-pkg-pcre)
    + libpcre++-dev
- [gouuid](https://github.com/nu7hatch/gouuid)
- [greenery](https://github.com/ferno/greenery)
- [Python 2.7.6](https://www.python.org/download/releases/2.7.6/)
- [gocc](https://code.google.com/p/gocc/)


Additional
---------------

The python scripts (`scripts/generate-*.py`) are used for testing. The tool does not depend on them.
The scripts depend on [xeger](https://pypi.python.org/pypi/xeger) (it can also be found in [rstr](https://pypi.python.org/pypi/rstr/2.1.2)).
