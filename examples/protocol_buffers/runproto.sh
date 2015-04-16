go build -o sublimegen -a ../../src/
./sublimegen -fileTypes "proto" -name "ProtoBuff" -scopeName "source.walt" -source languagefiles/proto.bnf -scopes languagefiles/scopes.json
