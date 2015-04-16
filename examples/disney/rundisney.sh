go build -o sublimegen -a ../../src/
./sublimegen -fileTypes "walt" -name "Walter" -scopeName "source.walt" -source languagefiles/asm.bnf -scopes languagefiles/scopes-new.json
