rm -rf Calc; go clean; go build; ./sublimegen -fileTypes "calc" -name "Calc" -scopeName "source.calc" -source calc_languagefiles/calc.bnf -scopes calc_languagefiles/scopes.json -orderregex 1
