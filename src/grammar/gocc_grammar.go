package grammar

import (
        "code.google.com/p/gocc/frontend/scanner"
        "code.google.com/p/gocc/frontend/token"
        "code.google.com/p/gocc/ast"
        "code.google.com/p/gocc/frontend/parser"
        "github.com/AdeebNqo/sublimegen/src/model"
        )

type GoccGrammar struct{
        SrcBuffer []byte

        grammar *ast.Grammar;
}
func NewGoccGrammar(srcBuffer []byte) (*GoccGrammar, error){
        scanner := &scanner.Scanner{}

        g := &GoccGrammar{SrcBuffer: srcBuffer}

        scanner.Init(srcBuffer, token.FRONTENDTokens)
        parser := parser.NewParser(parser.ActionTable, parser.GotoTable, parser.ProductionsTable, token.FRONTENDTokens)
        grammar, err := parser.Parse(scanner)

        g.grammar = grammar.(*ast.Grammar);

        return g, err
}
func (grammar GoccGrammar) ReadTokens() []model.Token{
        return nil
}
func (grammar GoccGrammar) ReadProductions() []model.Production{
        return nil
}
func (grammar GoccGrammar) IsLexPartEmpty() bool{
        return grammar.grammar.LexPart==nil
}
func (grammar GoccGrammar) IsSyntaxPartEmpty() bool{
        return grammar.grammar.SyntaxPart==nil
}

func (grammar GoccGrammar) DoesLexPartHaveProductions() bool{
        return grammar.grammar.SyntaxPart.ProdList!=nil
}
func (grammar GoccGrammar) DoesSyntaxPartHaveProductions() bool{
        return grammar.grammar.LexPart.ProdList.Productions!=nil
}
