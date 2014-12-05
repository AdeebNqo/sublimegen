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
    "strings"
	"reflect"
    "github.com/AdeebNqo/sublimegen/repository"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source","defaultinput", "the bnf file for the language you want to highlight.")

/*


*/

func stripliteral(somelit string) (retval string){
    if somelit != "" {
        somelit = somelit[1:len(somelit)-1]
    }
    retval = somelit
    return
}

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
    
    //retrieving the toekens and productions from the
    //the grammar
    grammarX := grammar.(*ast.Grammar)
    
    tokendefs := grammarX.LexPart.TokDefs
    
    fmt.Println(reflect.TypeOf(tokendefs))
    for key,value := range tokendefs{ //the key is the name that appears on the left hand side, value is the right hand side
        
        //extracting the name of the of the type from the key, remove the "lit"/"var" section
        startpos := 0
        if strings.HasPrefix(key,"_") {
            startpos = 1
        }
        endpos := strings.LastIndex(key, "_")
        if endpos==-1{
            endpos = len(key)-1
        }
        
        //

        name := key[startpos:endpos]
        
        fmt.Println(name)
        
        termalternatives := value.LexPattern().Alternatives
        for _,val := range termalternatives{
            for _, term:= range val.Terms{
                fmt.Println(term)
                fmt.Println(reflect.TypeOf(term))
            }
            fmt.Println()
            fmt.Println()
        }
    }
    
    /*

    Processing productions

    */
    productions := grammarX.LexPart.ProdList.Productions
    for _,prod  := range productions {
        prodid := prod.Id()
        
        if strings.HasPrefix(prodid,"_"){
        /*
        
        This if statement will only capture the what the gocc user guide defines as "regular definitions" (see pg 24)
        -Processing regular definitions-
    
        */
            /*prodregex := prod.LexPattern()
            fmt.Println("id and regex:")
            fmt.Println(prodid)
            
            alternatives := prodregex.Alternatives
            numalternatives := cap(alternatives)
            fmt.Println(fmt.Sprintf("there are/is %v alternative(s).", numalternatives))
            
            if (numalternatives == 1){
                val := alternatives[0]
                fmt.Println(val)
                fmt.Println(reflect.TypeOf(val))
                fmt.Println()
                fmt.Println()
            }else{
                
                ValueLoop:
                    for _,val := range alternatives{

                        //for each alternative, we need to look at the terms
                        altterms := val.Terms
                        fmt.Println(fmt.Sprintf("val: %v",val))
                        for _,term := range altterms{
                            fmt.Println(fmt.Sprintf("term: %v",term))
                            switch term.(type){
                                case *ast.LexRegDefId:{
                                    fmt.Println("*ast.LexRegDefId") //include - means this is somthing like "int_var" 
                                    break
                                }
                                case *ast.LexGroupPattern:{
                                    fmt.Println("*ast.LexGroupPattern")//compund regex with round braces
                                    break
                                }
                                case *ast.LexCharLit:{
                                    fmt.Println("*ast.LexCharLit")//literal char
                                    break
                                }
                                case *ast.LexRepPattern:{
                                    fmt.Println("*ast.LexRepPattern")//compound regex, surrounded by curly braces
                                    break   
                                }
                                case *ast.LexOptPattern:{
                                    fmt.Println("*ast.LexOptPattern")// surrounded by square braces
                                    break   
                                }
                                case *ast.LexCharRange:{
                                    fmt.Println("*ast.LexCharRange")// range of characters
                                    break   
                                }
                                default:{
                                    fmt.Println(fmt.Sprintf("type %v is unrecognized, ignoring %v",reflect.TypeOf(term), val))
                                    break ValueLoop
                                }
                            }
                        }
                    }
                fmt.Println()
                fmt.Println()
            }*/
        }else if strings.HasPrefix(prodid,"!"){
        /*
        
        This if statement will only capture the what the gocc user guide defines as "ignored token identifiers" (see pg 24)
        -Processing ignored token identifiers-
    
        */
            
        }else{
            //fmt.Println("Else")
            //fmt.Println(prod)
            //fmt.Println()
            //fmt.Println()
        }
        
    }
    
    //constructing the syntax highlighting file for sublime text
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
    //generating uuid for syntax highlighting file
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
