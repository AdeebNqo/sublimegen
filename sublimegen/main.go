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
    //"strings"
    "reflect"
    "github.com/AdeebNqo/sublimegen/repository"
    "container/list"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source","defaultinput", "the bnf file for the language you want to highlight.")

/*

Method for strip the start and end (') characters from a token

*/

func stripliteral(somelit string) (retval string){
    if somelit != "" {
        somelit = somelit[1:len(somelit)-1]
    }
    retval = somelit
    return
}

/*

Method for retrieving lex pattern

*/
func getItem(container *list.List, id string) (*repository.Repoitem){
    for val := container.Front(); val != nil; val = val.Next(){
        rval := val.Value.(*repository.Repoitem)
        if repository.GetRealname(rval)==id{
            return rval
        }
    }
    return nil
}

/*
function for obtaining regex from lexpatern
*/
func switchpattern(tokenlist *list.List, alternative interface{}) string{
    switch alternative.(type){
                    case *ast.LexRegDefId:{
                        castedterm := alternative.(*ast.LexRegDefId)
                        token := getItem(tokenlist ,castedterm.Id)
                        retrieveregex(tokenlist, token)
                        return fmt.Sprintf("%v",repository.Getregex(token))
                        //repository.Appendregex(sometoken, fmt.Sprintf("%v",repository.Getregex(token)))
                    }
                    case *ast.LexCharRange:{
                        tmpregex := "["+stripliteral(alternative.(*ast.LexCharRange).From.String())+"-"+stripliteral(alternative.(*ast.LexCharRange).To.String())+"]"
                        
                        //repository.Appendregex(sometoken, tmpregex)
                        return tmpregex
                    }
                    case *ast.LexCharLit:{
                        //repository.Appendregex(sometoken, stripliteral(alternative.(*ast.LexCharLit).String()))
                        return stripliteral(alternative.(*ast.LexCharLit).String())
                    }
                    case *ast.LexRepPattern:{
                        pattern2 := alternative.(*ast.LexRepPattern).LexPattern
                        tmpregex := ""
                        for index,val := range pattern2.Alternatives{
                            if index > 0{
                                tmpregex += "|"
                            }
                            for _,alternative := range val.Terms{
                                tmpregex += switchpattern(tokenlist,alternative)
                            }
                        }
                        return tmpregex
                    }
                    case *ast.LexGroupPattern:{
                        pattern2 := alternative.(*ast.LexGroupPattern).LexPattern
                        tmpregex := ""
                        for index,val := range pattern2.Alternatives{
                            if index > 0{
                                tmpregex += "|"
                            }
                            for _,alternative := range val.Terms{
                                tmpregex += switchpattern(tokenlist,alternative)
                            }
                        }
                        return tmpregex
                    }
                    case *ast.LexOptPattern:{
                        pattern2 := alternative.(*ast.LexOptPattern).LexPattern
                        tmpregex := "("
                        for index,val := range pattern2.Alternatives{
                            if index > 0{
                                tmpregex += "|"
                            }
                            for _,alternative := range val.Terms{
                                tmpregex += switchpattern(tokenlist,alternative)
                            }
                        }
                        tmpregex += ")?"
                        return tmpregex
                    }
                    case *ast.LexDot:{
                        return "."
                    }
                    default:{
                        fmt.Println("in default, type: ", reflect.TypeOf(alternative), " value: ", alternative)
                        return ""
                    }
                }
}

/*
function for expanding regex
*/
func retrieveregex(tokenlist *list.List, sometoken *repository.Repoitem){
    if repository.Isregexempty(sometoken){
        pattern := repository.GetRighthandside(sometoken)
        for index,val := range pattern.Alternatives{
            if index > 0{
                repository.Appendregex(sometoken,"|")
            }
            for _,alternative := range val.Terms{
                repository.Appendregex(sometoken, switchpattern(tokenlist, alternative))
            }
        }
    }
}


/*

Function for creating pattern entry from lexpattern

*/
func createpattern(somepattern *ast.LexPattern) (int, string, *list.List){
    
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

    //retrieving the tokens and productions from the
    //the grammar
    grammarX := grammar.(*ast.Grammar)
    
    tmpscope := fmt.Sprintf("keyword.control.%v",*fileTypes) //default scope that will be used in the meantime
    var repoitems *list.List
    repoitems = list.New()
    repoitems.Init()

    /*

    Processing token definitions

    */
    tokendefs := grammarX.LexPart.TokDefs //list of token definitions
    for key,value := range tokendefs{ //the key is the name that appears on the left hand side, value is the right hand side

        //creating an object that will convert the token to the appropriate item for the json patterns field
        patternobj,err := repository.NewRepoItem(key)

        if err != nil{
            //ignoring token
            fmt.Println(fmt.Sprintf("could not process %v. reason: %v",key, err))
            break
        }
        repository.SetRighthandside(patternobj,value.LexPattern())
        repository.SetScope(patternobj, tmpscope)
        repoitems.PushBack(patternobj)
    }

    /*

    Processing productions

    */
    productions := grammarX.LexPart.ProdList.Productions
    for _,prod  := range productions {
        prodid := prod.Id()


        //creating an object that will convert the production/token to the appropriate item for the json patterns field
        patternobj,err := repository.NewRepoItem(prodid)

        if err != nil{
            //ignoring token
            fmt.Println(fmt.Sprintf("could not process %v. reason: %v",prodid, err))
            break
        }
        repository.SetRighthandside(patternobj,prod.LexPattern())
        repository.SetScope(patternobj, tmpscope)
        repoitems.PushBack(patternobj)
    }

    //constructing the syntax highlighting file for sublime text
	jsonhighlight := `
		{ "name": "%v",
		  "scopeName": "%v",
		  "fileTypes": ["%v"],
		  "patterns": [
		  	%v
                    ],
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
        for listitem := repoitems.Front(); listitem != nil; listitem = listitem.Next() {
            listitemwithtype := listitem.Value.(*repository.Repoitem)
            repository.SetScope(listitemwithtype, tmpscope)

            alternatives := repository.GetRighthandside(listitemwithtype).Alternatives

            regex := ""
            group := 0
            groups := list.New()
            groups.Init()

            for index,val := range alternatives{
                //adding or in regex if there are alternatives
                if index>0{
                    regex += "|"
                }
                //processing elements
                for _, term:= range val.Terms{
                    switch term.(type){
                        case *ast.LexCharLit:{
                            regex += stripliteral(term.String())
                        }
                        case *ast.LexRegDefId:{
                            castedterm := term.(*ast.LexRegDefId)
                            token := getItem(repoitems ,castedterm.Id)

                            if repository.Isregexempty(token){
                                retrieveregex(repoitems, token) //expand regex
                            }
                            //get regex, assign group if we are not working with comment.
                            group += 1
                            regex += "("+repository.Getregex(token)+")"
                            groups.PushBack(repository.GetScope(token))
                        }
                        default:{
                            fmt.Println(fmt.Sprintf("ignored: <%v, %v>",term, reflect.TypeOf(term)))
                        }
                    }
                }
            }

            //setting regex
            if repository.Isregexempty(listitemwithtype){
                repository.Setregex(listitemwithtype, regex)
            }
            fmt.Println(repository.GetRealname(listitemwithtype))
            fmt.Println(fmt.Sprintf("%v, %v",regex, group))
            fmt.Println()
            fmt.Println()
        }
        
		result := fmt.Sprintf(jsonhighlight, *name, *scope, *fileTypes, repositoryfield, u)
        
		//1. save result in a JSON file.

		//2. convert result to a plist file and save it.

		fmt.Println(result)
        
	}
}