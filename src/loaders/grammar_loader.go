package loaders

import (
        "github.com/AdeebNqo/sublimegen/src/grammar"
        )

func GetGrammar(srcBuffer []byte) (grammar.IGrammar, error){
        return grammar.NewGoccGrammar(srcBuffer);
}
