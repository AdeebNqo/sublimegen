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
    "container/list"
    "encoding/json"
    "regexp"
    "strconv"
    "os/exec"
    "github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre" //(documentation: https://godoc.org/github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre)
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source","defaultinput", "the bnf file for the language you want to highlight.")

var repoitems *list.List

/*

Method for retrieving a repo item from a list.
Should be used in cases where the regex of a regular
definition is neeeded.

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

Method for strip the start and end (') characters from a token

*/

func stripliteral(somelit string) string{
    if somelit != "" {
        somelit = somelit[1:len(somelit)-1]
    }
    return somelit
}

/*

function for reall retrieving regex from individual lexterm items

return a string, the regex
*/
func reallygetregex(lexterm interface{}) string{
    switch lexterm.(type){
        case *ast.LexCharLit:{
            term := lexterm.(*ast.LexCharLit)
            termasstring := stripliteral(term.String())
            
            //The outside character classes are .^$*+?()[{\|
            //The inside character classes are ^-]\
            
            if termasstring=="/"{
                return "\\/"
            }else if termasstring=="$"{
                return "\\$"
            }else if termasstring=="*"{
                return "\\*"
            }else if termasstring=="["{
                return "\\["
            }else if termasstring=="]"{
                return "\\]"
            }else if termasstring=="\\"{
                return "\\\\"
            }else if termasstring==" "{
                return "[ ]"
            }
            
            return termasstring
            /*if len(termasstring)==2{
                fmt.Println("length is two. ")
                return stripliteral(termasstring) //only cater for the \t, \n, etc
            }else{
                return regexp.QuoteMeta(stripliteral(termasstring))
            }*/
        }
        case *ast.LexCharRange:{
            term := lexterm.(*ast.LexCharRange)
            retval := fmt.Sprintf("[%v-%v]",reallygetregex(term.From),reallygetregex(term.To))
            //fmt.Println("hello",retval) //debug
            return retval
        }
        case *ast.LexGroupPattern:{
            term := lexterm.(*ast.LexGroupPattern)
            //fmt.Println("pattern:",term.LexPattern) //debug
            retval := "("
            for index,lexalt := range term.LexPattern.Alternatives{
                if index>0{
                    retval += "|"
                }
                retval +=  getregex(lexalt)
            }
            retval += ")"
            //fmt.Println("retval:",retval) //debug
            return retval
        }
        case *ast.LexOptPattern:{
            term := lexterm.(*ast.LexOptPattern)

            alternatives := term.LexPattern.Alternatives
            //TODO; Remove round braces if the optional term is a terminal.
            retval := "("
            for index,lexalt := range alternatives{
                if index>0{
                    retval += "|"
                }
                retval += getregex(lexalt)
            }
            retval += ")?"
            //fmt.Println(retval) //debug
            return retval   
        }
        case *ast.LexRepPattern:{
            term := lexterm.(*ast.LexRepPattern)
            
            alternatives := term.LexPattern.Alternatives
   
             //TODO; Remove round braces if the optional term is a terminal.
            retval := "("
            for index,lexalt := range alternatives{
                if index>0{
                    retval += "|"
                }
                retval += getregex(lexalt)
            }
            retval += ")*"
            //fmt.Println(retval) //debug
            return retval
        }
        case *ast.LexDot:{
            return "."
        }
        case *ast.LexRegDefId:{
            term := lexterm.(*ast.LexRegDefId)
            
            //TODO: do not fetch rval, use a pointer
            for val := repoitems.Front(); val != nil; val = val.Next(){
                rval := val.Value.(*repository.Repoitem)
                if repository.GetRealname(rval)==term.Id{
                    if repository.Isregexempty(rval){
                
                        alternatives := repository.GetRighthandside(rval).Alternatives
                        retval := ""
                        for index,lexalt := range alternatives{
                            if index>0{
                                retval += "|"
                            }
                            retval += getregex(lexalt)
                        }
                        repository.Setregex(rval,retval)
                        return retval
                        
                    }else{
                        return repository.Getregex(rval)
                    }
                }
            }
        }
        default:{
            //fmt.Println("type:",reflect.TypeOf(lexterm), "term: ",lexterm)
            return "A"
        }
    }
    return ""
}

/*

function for retrieving regex from individual lexalt item

return a string, the regex
*/
func getregex(lexitem *ast.LexAlt) string{
    regex := ""
    for _,term := range lexitem.Terms{
        regex += reallygetregex(term)
    }
    return regex
}
/*

Function for unravelling a pattern to obtain it's regex

*/
func constructregexandfillgroups(alternatives []*ast.LexAlt, grouplist *list.List) (*list.List,string){
    fmt.Println("bnf:",alternatives) //debug
    regex := ""
    for _,lexitem := range alternatives{
        regex += getregex(lexitem)
    }
    fmt.Println("json:",regex) //debug
    return list.New().Init(),regex
}


//the following structs are used to produce a json file
type JSONSyntax struct{
    Name string `json:"name"`
    ScopeName string `json:"scopeName"`
    FileTypes []string `json:"fileTypes"`
    Patterns []PatternEntry `json:"patterns,omitempty"`
    Uuid string `json:"uuid"`
}
type PatternEntry struct{
    Match string `json:"match,omitempty"`
    Name string `json:"name,omitempty"`
    Captures map[string]CaptureEntryName `json:"captures,omitempty"`
}
type CaptureEntryName struct{
    Name string `json:"name,omitempty"`
}

func main() {
	flag.Parse() //parsing commandline flags

    //reading in the provided bnf file and parsing it.
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

    //loading tokens and scopes
    defaultscope := fmt.Sprintf("source.%v",*fileTypes) //default scope
    type config map[string]string
    var data config
    file, _ := ioutil.ReadFile("scopes.json")
    err = json.Unmarshal(file, &data)
    if err!=nil{
        fmt.Printf("Err(%s) : cannot parse json file with scopes", err)
		os.Exit(1)
    }

    //retrieving the tokens and productions from the
    //the grammar
    grammarX := grammar.(*ast.Grammar)

    //instantiating the list that will contain all the repoitems, that is, pattern field entries.
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
        somescope := data[repository.GetRealname(patternobj)]
        if somescope!=""{
            repository.SetScope(patternobj,somescope)
        }else{
            repository.SetScope(patternobj, defaultscope)
        }
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
        somescope := data[repository.GetRealname(patternobj)]
        if somescope!=""{
            repository.SetScope(patternobj,somescope)
        }else{
            repository.SetScope(patternobj, defaultscope)
        }
        
        repoitems.PushBack(patternobj)
    }

    //generating uuid for syntax highlighting file
	u, err := uuid.NewV4()
	if (err!=nil){
        //it was not possible to generated uuid, quiting...

		fmt.Println("Could not generate uuid.")
		os.Exit(1)
	}else{
        //genating patterns since uuid has been successfully generated

        patternarray := make([]PatternEntry,1)
        //0. Generate repository field from bnf file
        for listitem := repoitems.Front(); listitem != nil; listitem = listitem.Next() {
            listitemwithtype := listitem.Value.(*repository.Repoitem)

            alternatives := repository.GetRighthandside(listitemwithtype).Alternatives
            
            regex := ""
            groups := list.New()
            groups.Init()
            
            groups, regex = constructregexandfillgroups(alternatives, groups) //we are extracting the regex for the 

            //testing if regex is okay
            _, compileerr := pcre.Compile(regex,0)
            if compileerr!=nil{
                fmt.Println("err:",compileerr)
                fmt.Println()
            }
            
            //setting regex
            if repository.Isregexempty(listitemwithtype){
                repository.Setregex(listitemwithtype, regex)
            }


            //In the following lines, I am creating the "patterns" field for the json string declared above.
            //the final string will create the json file which will further be converted to plist. In particular,
            // I am creating the items (match and name, alongside the neccessary groups) which will be contained in "patterns" array.

            captureindex := 1
            numberofgroups := groups.Len()
            capturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

            if numberofgroups != 0{
                for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next(){
                    val := listitemX.Value.(string)
                    lastindex := strings.LastIndex(val, "|")
                    if lastindex > -1 {
                        //captureregex := val[0:lastindex]

                        capturename := val[lastindex+1:len(val)]
                        capturesmap[strconv.Itoa(captureindex)] = CaptureEntryName{Name:capturename}
                    }
                    captureindex += 1
                }

                //creating pattern entry
                patternentry := PatternEntry{Match:regex,Name:repository.GetScope(listitemwithtype),Captures:capturesmap}
                patternarray = append(patternarray, patternentry)
            }
        }
        
		//result := fmt.Sprintf(jsonhighlight, *name, *scope, *fileTypes, repositoryfield, u)
        jsonsyntaxobj := JSONSyntax{Name:*name, ScopeName:*scope, FileTypes:strings.Split(*fileTypes,","), Patterns:patternarray, Uuid:u.String()}
        jsonsyntaxobj_result, err := json.MarshalIndent(jsonsyntaxobj,"", "  ")
        if err!=nil{
            fmt.Println("we have a problem marshalling output.")
            os.Exit(1)
        }else{
            err := ioutil.WriteFile(fmt.Sprintf("%v.tmLanguage.json", *name), jsonsyntaxobj_result, 0644)
             if err!=nil{
                fmt.Println("(Error): we have a problem writing to file.")
                fmt.Println("(Reason):", err)
                os.Exit(1)
             }
        }
        
		//convert resulting json to a plist file and save it.
		err = exec.Command("python", "convertor.py", fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v.tmLanguage", *name)).Run()
        if err!=nil{
            fmt.Println(fmt.Sprintf("(Error): Could not convert json to plist.\n(Reason): %v",err))
            os.Exit(1)
        }
        //moving the files into a folder with the name provided as cmdline arg
        directoryexists := true
        if _, err := os.Stat(*name); err != nil {
            if os.IsNotExist(err) {
                        directoryexists = false
                    }
        }
        if directoryexists{
            err := os.RemoveAll(*name)
            if err!=nil{
                fmt.Println("(Error)Cannot remove old directory")
                fmt.Println("(Reason):",err)
            }
        }
        
        err0 := os.Mkdir(*name, 0775)
        if err0!=nil{
            fmt.Println("(Error): Could not create folder for storing generated files")
            fmt.Println("(Reason):", err0)
            os.Exit(1)
        }
        err1 := os.Rename(fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v/%v.tmLanguage.json", *name, *name))
        err2 := os.Rename(fmt.Sprintf("%v.tmLanguage", *name), fmt.Sprintf("%v/%v.tmLanguage", *name, *name))
        if err1!=nil || err2!=nil {
            fmt.Println("(Error): Could not move files")
            fmt.Println("(Reason):", err1,"and/or", err2)
            os.Exit(1)
        }
	}
    os.Exit(0)
}







///=====================================================================================================================================
///=====================================================================================================================================
///=====================================================================================================================================
///All the code below has been abandonded.
///=====================================================================================================================================
///=====================================================================================================================================
///=====================================================================================================================================



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
                        tmpregex += ")" //i just removed a question mark here.
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

return values: group, regex and groups
*/
func createpattern(group int, regex string, groups *list.List, repoitems *list.List, term interface{}) (int, string, *list.List){
    fmt.Println("term:",term,",type:",reflect.TypeOf(term)) //debug
    /*switch term.(type){
        case *ast.LexCharLit:{
            termX := term.(*ast.LexCharLit)
            return group, stripliteral(termX.String()), groups
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

            groups.PushBack(repository.Getregex(token)+"|"+repository.GetScope(token))
            return group, regex, groups
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

            groups.PushBack(repository.Getregex(token)+"|"+repository.GetScope(token))
            return group, regex, groups
        }
        case *ast.LexOptPattern:{
            pattern2 := term.(*ast.LexOptPattern).LexPattern
            tmpregex := "("
            for index,val := range pattern2.Alternatives{
                if index > 0{
                    tmpregex += "|"
                }
                for _,alternative := range val.Terms{
                    tmpregex += switchpattern(repoitems, alternative)
                }
            }
            tmpregex += ")" //removed question mark at the end because python does not like nested optional quantifiers
            group += 1
            regex += tmpregex
            groups.PushBack(tmpregex+"|keyword.control.bnf") //for consistency
            return group, regex, groups
        }
        case *ast.LexCharRange:{
            pattern2 := term.(*ast.LexCharRange)
            regex +=  "["+stripliteral(pattern2.From.String())+"-"+stripliteral(pattern2.To.String())+"]"
            return group, regex, groups
        }
        case *ast.LexGroupPattern:{
            pattern2 := term.(*ast.LexGroupPattern).LexPattern
            for index,val := range pattern2.Alternatives{
                //adding or in regex if there are alternatives
                if index>0{
                    regex += "|"
                }
                //processing elements
                for _, term2:= range val.Terms{
                    group, regex, groups = createpattern(group, regex, groups, repoitems, term2)
                    return group, regex, groups
                }
            }
        }
        case *ast.LexDot:{
            regex += "."
            return group, regex, groups
        }
        case *ast.LexRepPattern:{
            pattern2 := term.(*ast.LexRepPattern).LexPattern
            //fmt.Println(pattern2) //debug
            for index,val := range pattern2.Alternatives{
                //adding or in regex if there are alternatives
                if index>0{
                    regex += "|"
                }
                //processing elements
                for _, term2:= range val.Terms{
                    group, regex, groups = createpattern(group, regex, groups, repoitems, term2)
                    //regex = fmt.Sprintf("(%v)*",regex)
                    //fmt.Println(regex) //debug
                    return group, regex, groups
                }
            }
        }
        default:{
            fmt.Println(reflect.TypeOf(term))
        }
    }*/
    return 0,"",nil
}

/*

Method for strip the start and end (') characters from a token

*/

/*func stripliteral(somelit string) (retval string){
    if somelit != "" {
        somelit = somelit[1:len(somelit)-1]
    }
    retval = escape(somelit)
    //retval = somelit
    return
}
*/


/*

function for escaping char lits
for a regex

*/
func escape(somechar string) string{
    //fmt.Println("value:", somechar, "length:",len(somechar)) //debug
    if somechar==" "{
        return "[ ]"
    }else if somechar=="\n"{
        return "$"
    }else if len(somechar)!=1{
        return somechar
    }else{
        return regexp.QuoteMeta(somechar)
    }
}