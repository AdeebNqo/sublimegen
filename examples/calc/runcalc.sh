go build -o sublimegen -a ../../src/
./sublimegen -fileTypes "calc" -name "Calc" -scopeName "source.calc" -source languagefiles/calc.bnf -scopes languagefiles/scopes.json
