package production

import (
        "code.google.com/p/gocc/ast"
        )

type GoccProduction struct{
        lexProduction *ast.LexProduction;
        syntaxProduction *ast.SyntaxProd;
}

func NewGoCCProduction(someProduction interface{}) (*GoccProduction, error){
        var goccProd GoccProduction

        switch someProduction.(type){
                case *ast.LexProduction:{
                        castedProduction := someProduction.(*ast.LexProduction)
                        goccProd = GoccProduction{lexProduction:castedProduction}
                }
                case *ast.SyntaxProd:{
                        castedProduction := someProduction.(*ast.SyntaxProd)
                        goccProd = GoccProduction{syntaxProduction:castedProduction}
                }
        }

        return &goccProd, nil
}

func (someProd GoccProduction) ReadLeftHand() string{
        var name string

        if (someProd.lexProduction == nil){
                //syntax productions being used
                name = (*someProd.syntaxProduction).Id
        }else{
                //lex productions being used
                name = (*someProd.lexProduction).Id()
        }
        return name
}
func (someProd GoccProduction) ReadRightHand() interface{}{
        if (someProd.lexProduction != nil){
                return (*someProd.lexProduction).LexPattern()
        }
        return (*someProd.syntaxProduction).Body
}
