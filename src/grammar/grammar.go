package grammar

import (
        "github.com/AdeebNqo/sublimegen/src/production"
        )
type IGrammar interface{
        ReadLexPartProductions() []production.IProduction
        ReadSyntaxPartProductions() []production.IProduction

        IsLexPartEmpty() bool
        IsSyntaxPartEmpty() bool

        DoesLexPartHaveProductions() bool
        DoesSyntaxPartHaveProductions() bool
}
