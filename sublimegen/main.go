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
	"io/ioutil"
    "code.google.com/p/gocc/ast"
    
	"os"
	"flag"
	"fmt"
	"github.com/nu7hatch/gouuid"
    "strings"
    "github.com/AdeebNqo/sublimegen/repository"
    "container/list"
    "encoding/json"
    "os/exec"
    "sort"
    "github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre" //(documentation: https://godoc.org/github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre)
    "github.com/AdeebNqo/sublimegen/logger"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source","defaultinput", "the bnf file for the language you want to highlight.")
var doregexorder = flag.Int("orderregex",1,"Program to attempt to order regexes in file. 0 for no, 1 for yes")
var verbose = flag.Int("verbose",1,"Output status and other progress information. 0 for no, 1 for yes")
var mylogger = logger.Init(os.Stdout, os.Stdout, os.Stderr)

var repoitems *list.List
var defaultscope string


func main() {
	flag.Parse() //parsing commandline flags

    //reading in the provided bnf file and parsing it.
	scanner := &scanner.Scanner{}
	srcBuffer, err := ioutil.ReadFile(*source)
	if err != nil {
        mylogger.Err(fmt.Sprintf("Cannot read file because %v",err))
		os.Exit(1)
	}
	scanner.Init(srcBuffer, token.FRONTENDTokens)
	parser := parser.NewParser(parser.ActionTable, parser.GotoTable, parser.ProductionsTable, token.FRONTENDTokens)
	grammar, err := parser.Parse(scanner)
	
    if err != nil {
        mylogger.Err(fmt.Sprintf("Parse error: %v",err))
		os.Exit(1)
	}

    //loading tokens and scopes
    defaultscope = fmt.Sprintf("source.%v",*fileTypes) //default scope
    type config map[string]string
    var data config
    file, _ := ioutil.ReadFile("scopes.json")
    err = json.Unmarshal(file, &data)
    if err!=nil{
        mylogger.Err(fmt.Sprintf("Cannot parse json file with scopes because %v",err))
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
    /*tokendefs := grammarX.LexPart.TokDefs //list of token definitions
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
    }*/

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
            mylogger.Err(fmt.Sprintf("could not process %v. reason: %v",prodid, err))
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

    if *verbose==1{
        mylogger.Inform("Generating uuid for syntax highlighting file.")
    }
    //generating uuid for syntax highlighting file
	u, err := uuid.NewV4()
	if (err!=nil){
        //it was not possible to generated uuid, quiting...
        mylogger.Err("Could not generate uuid.")
		//fmt.Println("Could not generate uuid.")
		os.Exit(1)
	}else{
        mylogger.Inform("Finished generating uuid. Now processing bnf file...")
        //genating patterns since uuid has been successfully generated

        //patternarray := make([]PatternEntry,1)
        patternarray := make(patternarraytype,0)
        //0. Generate repository field from bnf file
        for listitem := repoitems.Front(); listitem != nil; listitem = listitem.Next() {
            listitemwithtype := listitem.Value.(*repository.Repoitem)

            realname := repository.GetRealname(listitemwithtype)
            alternatives := repository.GetRighthandside(listitemwithtype).Alternatives

            regex := constructregexandfillgroups(alternatives) //we are extracting the regex for the  
            
            //testing if regex is okay
            regp, compileerr := pcre.Compile(regex,0)
            if compileerr!=nil{
                //regex is compatile so skip it.
                mylogger.Err(compileerr.String())
                //fmt.Println("err:",compileerr)
                break
            }

            //setting regex
            if repository.Isregexempty(listitemwithtype){
                repository.Setregex(listitemwithtype, regex)
            }

            //determining if one should use begin and end
            usebeginandend, begin, middle, end:= determinebeginandend(regex)

            if usebeginandend{

                //sorting out captures for begin regex
                groups,_ := getgroups(begin , begin ,0, list.New().Init(), list.New().Init())
                numberofgroups := groups.Len()
                begincapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

                donotskip := true //variable to be used to skip the groups --- kinda a "hack"
                skippingscope := repository.GetScope(listitemwithtype)
                skippingfront := groups.Front()

                if numberofgroups==1{
                    if skippingfront==nil || skippingscope==defaultscope{
                        donotskip = false
                    }
                    skippingfrontvalue := skippingfront.Value.(string)
                    skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
                    donotskip = !(skippingtruefrontvalue==skippingscope)
                }
                //adding items to "begincaptures"
                if numberofgroups>0 && regp.Groups() !=0 && donotskip {
                    for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next(){
                        val := listitemX.Value.(string)
                        lastindex := strings.LastIndex(val, "|")
                        if lastindex > -1 {
                            scopename := val[0:lastindex]
                            scopenumber := val[lastindex+1:len(val)]
                            begincapturesmap[scopenumber] = CaptureEntryName{Name:scopename}
                        }
                    }
                }
                //----------------------------------------------------------------------------------------------

                //sorting out captures for end regex
                groups,_ = getgroups(end , end ,0, list.New().Init(), list.New().Init())
                numberofgroups = groups.Len()
                endcapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

                donotskip = true //variable to be used to skip the groups --- kinda a "hack"
                skippingscope = repository.GetScope(listitemwithtype)
                skippingfront = groups.Front()

                if numberofgroups==1{
                    if skippingfront==nil || skippingscope==defaultscope{
                        donotskip = false
                    }
                    skippingfrontvalue := skippingfront.Value.(string)
                    skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
                    donotskip = !(skippingtruefrontvalue==skippingscope)
                }
                //adding items to "begincaptures"
                if numberofgroups>0 && regp.Groups() !=0 && donotskip {
                    for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next(){
                        val := listitemX.Value.(string)
                        lastindex := strings.LastIndex(val, "|")
                        if lastindex > -1 {
                            scopename := val[0:lastindex]
                            scopenumber := val[lastindex+1:len(val)]
                            endcapturesmap[scopenumber] = CaptureEntryName{Name:scopename}
                        }
                    }
                }
                //---------------------------------------------------------------------------------------------

                //adding middle parts
                groups,_ = getgroups(middle , middle ,0, list.New().Init(), list.New().Init())
                numberofgroups = groups.Len()
                middlecapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

                donotskip = true //variable to be used to skip the groups --- kinda a "hack"
                skippingscope = repository.GetScope(listitemwithtype)
                skippingfront = groups.Front()

                if numberofgroups==1{
                    if skippingfront==nil || skippingscope==defaultscope{
                        donotskip = false
                    }
                    skippingfrontvalue := skippingfront.Value.(string)
                    skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
                    donotskip = !(skippingtruefrontvalue==skippingscope)
                }
                //adding items to "begincaptures"
                if numberofgroups>0 && regp.Groups() !=0 && donotskip {
                    for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next(){
                        val := listitemX.Value.(string)
                        lastindex := strings.LastIndex(val, "|")
                        if lastindex > -1 {
                            scopename := val[0:lastindex]
                            scopenumber := val[lastindex+1:len(val)]
                            middlecapturesmap[scopenumber] = CaptureEntryName{Name:scopename}
                        }
                    }
                }
                middlearray := make(patternarraytype,0)
                middlearray = append(middlearray, PatternEntry{Match:middle, Captures:middlecapturesmap})
                //--------------------------------------------------------------------------------------------


                //creating pattern entry
                patternentry := PatternEntry{Begin:begin, End:end,Name:repository.GetScope(listitemwithtype), EndCaptures:endcapturesmap, MorePatterns:middlearray}
                patternarray = append(patternarray, patternentry)
            }else{

                if !strings.HasPrefix(realname, "_"){
                    //getting groups
                    groups,_ := getgroups(regex , regex ,0, list.New().Init(), list.New().Init())

                    //In the following lines, I am creating the "patterns" field for the json string declared above.
                    //the final string will create the json file which will further be converted to plist. In particular,
                    // I am creating the items (match and name, alongside the neccessary groups) which will be contained in "patterns" array.

                    numberofgroups := groups.Len()
                    capturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

                    donotskip := true //variable to be used to skip the groups --- kinda a "hack"
                    skippingscope := repository.GetScope(listitemwithtype)
                    skippingfront := groups.Front()


                    if numberofgroups==1{
                        if skippingfront==nil || skippingscope==defaultscope{
                            donotskip = false
                        }
                        skippingfrontvalue := skippingfront.Value.(string)
                        skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
                        donotskip = !(skippingtruefrontvalue==skippingscope)
                    }

                    //adding items to "captures"
                    if numberofgroups>0 && regp.Groups() !=0 && donotskip {
                        for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next(){
                            val := listitemX.Value.(string)
                            lastindex := strings.LastIndex(val, "|")
                            if lastindex > -1 {
                                scopename := val[0:lastindex]
                                scopenumber := val[lastindex+1:len(val)]
                                capturesmap[scopenumber] = CaptureEntryName{Name:scopename}
                            }
                        }
                    }

                    //creating pattern entry
                    patternentry := PatternEntry{Match:regex,Name:repository.GetScope(listitemwithtype),Captures:capturesmap}
                    patternarray = append(patternarray, patternentry)
                }
            }
        }
        
        if *verbose==1{
            mylogger.Inform("Finished processing bnf file.")
        }
        
        if *doregexorder==1{
            //sorting regexes
            mylogger.Inform("Sorting regexes...")
            sort.Sort(patternarray)
            mylogger.Inform("Done sorting!")
        }
        
        if *verbose==1{
            mylogger.Inform("Transforming syntax highlighting data to json...")
        }
        
        //marshaling output into proper json
        jsonsyntaxobj := JSONSyntax{Name:*name, ScopeName:*scope, FileTypes:strings.Split(*fileTypes,","), Patterns:patternarray, Uuid:u.String()}
        jsonsyntaxobj_result, err := json.MarshalIndent(jsonsyntaxobj,"", "  ")
        
        if err!=nil{
            if *verbose==1{
                mylogger.Err(fmt.Sprintf("Could not transform syntax highlighting data to json becase %v",err))
            }
            //fmt.Println("we have a problem marshalling output.")
            os.Exit(1)
        }else{
            if (*verbose==1){
                mylogger.Inform("done converting syntax highlighting data to json. Now saving json file...")
            }
            err := ioutil.WriteFile(fmt.Sprintf("%v.tmLanguage.json", *name), jsonsyntaxobj_result, 0644)
            if err!=nil{
                mylogger.Err(fmt.Sprintf("We a problem writing to file because %v",err))
                //fmt.Println("(Error): we have a problem writing to file.")
                //fmt.Println("(Reason):", err)
                os.Exit(1)
            }else{
                if (*verbose==1){
                    mylogger.Inform("json file saved.")
                }
            }
        }
        
        if *verbose==1{
            mylogger.Inform("Converting json to plist...")
        }
		//convert resulting json to a plist file and save it.
		err = exec.Command("python", "convertor.py", fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v.tmLanguage", *name)).Run()
        if err!=nil{
            mylogger.Err(fmt.Sprintf("(Error): Could not convert json to plist.\n(Reason): %v",err))
            //fmt.Println(fmt.Sprintf("(Error): Could not convert json to plist.\n(Reason): %v",err))
            os.Exit(1)
        }else{
            if *verbose==1{
                mylogger.Inform("Finished converting json to plist!")
            }
        }
        //moving the files into a folder with the name provided as cmdline arg
        directoryexists := true
        if _, err := os.Stat(*name); err != nil {
            if os.IsNotExist(err) {
                directoryexists = false
            }
        }

        //removing old directory with the same name
        if directoryexists{
            if (*verbose==1){
                mylogger.Inform("Found old directory with same name as target directory, deleting...")
            }
            err := os.RemoveAll(*name)
            if err!=nil{
                mylogger.Err(fmt.Sprintf("Cannot remove old directory because %v",err))
                //fmt.Println("(Error)Cannot remove old directory")
                //fmt.Println("(Reason):",err)
                os.Exit(1)
            }else{
                if (*verbose==1){
                    mylogger.Inform("Old directory with same name as target directory, deleted!")
                }
            }
        }
        
        //creating folder for syntax highlighting files
        err0 := os.Mkdir(*name, 0775)
        if err0!=nil{
            mylogger.Err(fmt.Sprintf("Could not create folder for storing generated files because %v",err0))
            //fmt.Println("(Error): Could not create folder for storing generated files")
            //fmt.Println("(Reason):", err0)
            os.Exit(1)
        }
        
        if (*verbose==1){
            mylogger.Inform("Moving files into new folder!")
        }
        //moving files into created folder
        err1 := os.Rename(fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v/%v.tmLanguage.json", *name, *name))
        err2 := os.Rename(fmt.Sprintf("%v.tmLanguage", *name), fmt.Sprintf("%v/%v.tmLanguage", *name, *name))
        if err1!=nil || err2!=nil {
            mylogger.Err(fmt.Sprintf("Could not move files because %v and/or %v",err1,err2))
            //fmt.Println("(Error): Could not move files")
            //fmt.Println("(Reason):", err1,"and/or", err2)
            os.Exit(1)
        }
	}
    
    if (*verbose==1){
        mylogger.Inform("Finished!")
    }
    os.Exit(0)
}