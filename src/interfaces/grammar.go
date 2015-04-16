type Grammar interface{
        ReadTokens() []Token
        ReadProductions() []Production

        isLexPartEmpty() bool
        isSyntaxPartEmpty() bool
}
