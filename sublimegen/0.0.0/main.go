/*

Copyright 2014 Zola Mahlaza <adeebnqo@gmail.com>
02 December 2014

This is the driver class. It is resposible for creating the
BNF loader/parser and generating the sublime text syntax highligting
file.

*/

package main

import (
	"code.google.com/p/gocc/frontend/scanner"
	"code.google.com/p/gocc/frontend/token"
	"code.google.com/p/gocc/frontend/parser"
    "code.google.com/p/gocc/ast"
	"io/ioutil"

	"os"
	"flag"
	"fmt"
	"github.com/nu7hatch/gouuid"
	//"reflect"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source","defaultinput", "the bnf file for the language you want to highlight.")

func main() {
	flag.Parse() //parsing commandline flags

    //walter
	scanner := &scanner.Scanner{}
	srcBuffer, err := ioutil.ReadFile(*source)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	scanner.Init(srcBuffer, token.FRONTENDTokens)
	parser := parser.NewParser(parser.ActionTable, parser.GotoTable, parser.ProductionsTable, token.FRONTENDTokens)
	grammar, err := parser.Parse(scanner)
	if err != nil {
		fmt.Printf("Parse error: %s\n", err)
		os.Exit(1)
	}
    
    //
    grammarX := grammar.(*ast.Grammar)
    tokendefs := grammarX.LexPart.TokDefs
    for key,value := range tokendefs{
        fmt.Println(key)
        fmt.Println(value)
    }
    
    productions := grammarX.LexPart.ProdList.Productions
    for _,prod  := range productions {
        fmt.Println(prod)
    }
    
	jsonhighlight := `

		{ "name": "%v",
		  "scopeName": "%v",
		  "fileTypes": [%v],
		  "repository": {
		  	%v
		  },
		  "uuid": "%v"
		}
	`
	u, err := uuid.NewV4()
	if (err!=nil){
		fmt.Println("Could not generate uuid.")
		os.Exit(1)
	}else{

		repositoryfield := "COMING..."
		//0. Generate repository field from bnf file

		result := fmt.Sprintf(jsonhighlight, *name, *scope, *fileTypes, repositoryfield, u)

		//1. save result in a JSON file.

		//2. convert result to a plist file and save it.

		fmt.Println(result)
	}
}
