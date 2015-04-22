package grammar

import (
        "code.google.com/p/gocc/frontend/scanner"
        "code.google.com/p/gocc/frontend/token"
        "code.google.com/p/gocc/ast"
        "code.google.com/p/gocc/frontend/parser"
        "github.com/AdeebNqo/sublimegen/src/production"

        "errors"
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

        if g.grammar==nil{
                return nil, errors.New("Cannot create grammar")
        }

        g.grammar = grammar.(*ast.Grammar);

        return g, err
}
func (grammar GoccGrammar) ReadLexPartProductions() []production.IProduction{
        goccProductions := make([]production.IProduction,1)
        for _, prod := range grammar.grammar.LexPart.ProdList.Productions{
                newGoccProduction ,_ := production.NewGoCCProduction(&prod)
                if newGoccProduction!=nil{
                        goccProductions = append(goccProductions, newGoccProduction)
                }
        }
        return goccProductions
}
func (grammar GoccGrammar) ReadSyntaxPartProductions() []production.IProduction{
        goccProductions := make([]production.IProduction,1)
        for _, prod := range grammar.grammar.SyntaxPart.ProdList{
                newGoccProduction ,_ := production.NewGoCCProduction(prod)
                if newGoccProduction!=nil{
                        goccProductions = append(goccProductions, newGoccProduction)
                }
        }
        return goccProductions
}
func (grammar GoccGrammar) IsLexPartEmpty() bool{
        return grammar.grammar == nil || grammar.grammar.LexPart==nil
}
func (grammar GoccGrammar) IsSyntaxPartEmpty() bool{
        return grammar.grammar == nil || grammar.grammar.SyntaxPart==nil
}

func (grammar GoccGrammar) DoesLexPartHaveProductions() bool{
        return grammar.grammar.SyntaxPart.ProdList!=nil
}
func (grammar GoccGrammar) DoesSyntaxPartHaveProductions() bool{
        return grammar.grammar.LexPart.ProdList.Productions!=nil
}
