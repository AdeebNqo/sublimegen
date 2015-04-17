package grammar

import (
        "github.com/AdeebNqo/sublimegen/src/model"
        )
type Grammar interface{
        ReadTokens() []model.Token
        ReadProductions() []model.Production

        IsLexPartEmpty() bool
        IsSyntaxPartEmpty() bool

        DoesLexPartHaveProductions() bool
        DoesSyntaxPartHaveProductions() bool
}
