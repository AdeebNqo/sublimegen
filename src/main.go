/*

Copyright 2014 Zola Mahlaza <adeebnqo@gmail.com>
02 December 2014

This is the driver class. It is resposible for creating the
BNF loader/parser and generating the sublime text syntax highligting
file.

*/

package main

import (
	"code.google.com/p/gocc/ast"
	"code.google.com/p/gocc/frontend/parser"
	"code.google.com/p/gocc/frontend/scanner"
	"code.google.com/p/gocc/frontend/token"
	"io/ioutil"

	"container/list"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/AdeebNqo/sublimegen/repository"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre" //(documentation: https://godoc.org/github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre)
	"github.com/nu7hatch/gouuid"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scopeName = flag.String("scopeName", "source.default", "This is the scope of the syntax file.")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")
var source = flag.String("source", "defaultinput", "the bnf file for the language you want to highlight.")
var scopesfile = flag.String("scopes", "scopes.json", "the json file containing the scope selectors.")
var doregexorder = flag.Int("orderregex", 1, "Program to attempt to order regexes in file. 0 for no, 1 for yes")
var verbose = flag.Int("verbose", 1, "Output status and other progress information. 0 for no, 1 for yes")

var repoitems *list.List
var defaultscope string

var errlog *log.Logger
var infolog *log.Logger

func main() {
	flag.Parse() //parsing commandline flags

	//reading in the provided bnf file and parsing it.
	scanner := &scanner.Scanner{}
	srcBuffer, err := ioutil.ReadFile(*source)

    //initializing logging objects
	errlog = log.New(os.Stdout, "Error: ", log.Ltime|log.Lshortfile)
	infolog = log.New(os.Stderr, "Info: ", log.Ltime|log.Lshortfile)

	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Cannot read file because %v", err))
	}
	scanner.Init(srcBuffer, token.FRONTENDTokens)
	parser := parser.NewParser(parser.ActionTable, parser.GotoTable, parser.ProductionsTable, token.FRONTENDTokens)
	grammar, err := parser.Parse(scanner)

	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Parse error: %v", err))
	}

	//loading tokens and scopes
	defaultscope = fmt.Sprintf("source.%v", *fileTypes) //default scope
	type config map[string]string
	var data config
	file, _ := ioutil.ReadFile(*scopesfile)
	err = json.Unmarshal(file, &data)
	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Cannot parse json file with scopes because %v", err))
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
	if grammarX.LexPart!=nil && grammarX.LexPart.ProdList!=nil && grammarX.LexPart.ProdList.Productions!=nil{
		productions := grammarX.LexPart.ProdList.Productions
		for _, prod := range productions {
			prodid := prod.Id()

			//creating an object that will convert the production/token to the appropriate item for the json patterns field
			patternobj, err := repository.NewRepoItem(prodid)
			if err != nil {
				//ignoring token
				infolog.Println(fmt.Sprintf("could not process %v. reason: %v", prodid, err))
				break
			}
			repository.SetRighthandside(patternobj, prod.LexPattern())
			somescope := data[repository.GetRealname(patternobj)]
			if somescope != "" {
				repository.SetScope(patternobj, somescope)
			} else {
				repository.SetScope(patternobj, defaultscope)
			}

			repoitems.PushBack(patternobj)
		}
	}

	/*

	   Processing the syntax part

	*/
	if grammarX.SyntaxPart!=nil{
		if grammarX.SyntaxPart.ProdList!=nil{
			for _, synprod := range grammarX.SyntaxPart.ProdList {
				for _, synsymb := range synprod.Body.Symbols {

					synprodname := synsymb.String()
					found := false
					for t := repoitems.Front(); t != nil; t = t.Next() {
						item := t.Value.(*repository.Repoitem)
						realname := repository.GetRealname(item)
						if realname == synprodname {
							found = true
							break
						} else if fmt.Sprintf("\"%v\"", realname) == synprodname {
							found = true
							break
						}
					}
					if found == false {
						if strings.HasPrefix(synprodname, "\"") && strings.HasSuffix(synprodname, "\"") {
							prodid := synprodname[1 : len(synprodname)-1]
							patternobj, err := repository.NewRepoItem(prodid)
							//_,err := repository.NewRepoItem(prodid)
							if err != nil {
								//ignoring token
								errlog.Fatalln(fmt.Sprintf("could not process %v. reason: %v", prodid, err))
								break
							}
							repository.SetRighthandside(patternobj, nil)
							somescope := data[repository.GetRealname(patternobj)]
							if somescope != "" {
								repository.SetScope(patternobj, somescope)
							} else {
								repository.SetScope(patternobj, defaultscope)
							}

							tmpprodid := ""
							for _, char := range prodid {
								tmpprodid += string(char)
							}
							prodid = tmpprodid
							if strings.ContainsAny(prodid, "abcdefghijklmnopqrstuvwxyz") && somescope != defaultscope {
								repoitems.PushBack(patternobj)
							}
						}
					}
				}
			}
		}
	}

	if *verbose == 1 {
		infolog.Println("Generating uuid for syntax highlighting file.")
	}
	//generating uuid for syntax highlighting file
	u, err := uuid.NewV4()
	if err != nil {
		//it was not possible to generated uuid, quiting...
		errlog.Fatalln("Could not generate uuid.")
	} else {
		infolog.Println("Finished generating uuid. Now processing bnf file...")
		//genating patterns since uuid has been successfully generated

		//patternarray := make([]PatternEntry,1)
		patternarray := make(patternarraytype, 0)
		//0. Generate repository field from bnf file
		for listitem := repoitems.Front(); listitem != nil; listitem = listitem.Next() {
			listitemwithtype := listitem.Value.(*repository.Repoitem)

			realname := repository.GetRealname(listitemwithtype)

            //fmt.Println(realname) //debug

			beforealternatives := repository.GetRighthandside(listitemwithtype)
			var regex string
			if beforealternatives != nil {
				alternatives := beforealternatives.Alternatives
				regex = constructregexandfillgroups(alternatives) //we are extracting the regex for the
			} else {
				for _, char := range realname {
					regex += Escape(string(char))
				}
			}

			// making sure that we do not match words that
			// are subwords
			if realname==regex || realname[1:]=regex{
				regex = fmt.Sprintf("(\\A|\\s)(%v)(\\z|\\s)?",regex)
			}
			//testing if regex is okay
			regp, compileerr := pcre.Compile(regex, 0)
			if compileerr != nil {
				//regex is not compatile so skip it.
				infolog.Println(compileerr.String())
				continue
			}

			//setting regex
			if repository.Isregexempty(listitemwithtype) {
				repository.Setregex(listitemwithtype, regex)
			}

			//determining if one should use begin and end
			usebeginandend, begin, middle, end := determinebeginandend(regex)

			if usebeginandend {

				//sorting out captures for begin regex
				groups, _ := getgroups(begin, begin, 0, list.New().Init(), list.New().Init())
				numberofgroups := groups.Len()
				begincapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

				donotskip := true //variable to be used to skip the groups --- kinda a "hack"
				skippingscope := repository.GetScope(listitemwithtype)
				skippingfront := groups.Front()

				if numberofgroups == 1 {
					if skippingfront == nil || skippingscope == defaultscope {
						donotskip = false
					}
					skippingfrontvalue := skippingfront.Value.(string)
					skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
					donotskip = !(skippingtruefrontvalue == skippingscope)
				}
				//adding items to "begincaptures"
				if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
					for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
						val := listitemX.Value.(string)
						lastindex := strings.LastIndex(val, "|")
						if lastindex > -1 {
							scopename := val[0:lastindex]
							scopenumber := val[lastindex+1:]
							begincapturesmap[scopenumber] = CaptureEntryName{Name: scopename}
						}
					}
				}
				//----------------------------------------------------------------------------------------------

				//sorting out captures for end regex
				groups, _ = getgroups(end, end, 0, list.New().Init(), list.New().Init())
				numberofgroups = groups.Len()
				endcapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

				donotskip = true //variable to be used to skip the groups --- kinda a "hack"
				skippingscope = repository.GetScope(listitemwithtype)
				skippingfront = groups.Front()

				if numberofgroups == 1 {
					if skippingfront == nil || skippingscope == defaultscope {
						donotskip = false
					}
					skippingfrontvalue := skippingfront.Value.(string)
					skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
					donotskip = !(skippingtruefrontvalue == skippingscope)
				}
				//adding items to "begincaptures"
				if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
					for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
						val := listitemX.Value.(string)
						lastindex := strings.LastIndex(val, "|")
						if lastindex > -1 {
							scopename := val[0:lastindex]
							scopenumber := val[lastindex+1:]
							endcapturesmap[scopenumber] = CaptureEntryName{Name: scopename}
						}
					}
				}

				//---------------------------------------------------------------------------------------------

				//adding middle parts
				groups, _ = getgroups(middle, middle, 0, list.New().Init(), list.New().Init())
				numberofgroups = groups.Len()
				middlecapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

				donotskip = true //variable to be used to skip the groups --- kinda a "hack"
				skippingscope = repository.GetScope(listitemwithtype)
				skippingfront = groups.Front()

				if numberofgroups == 1 {
					if skippingfront == nil || skippingscope == defaultscope {
						donotskip = false
					}
					skippingfrontvalue := skippingfront.Value.(string)
					skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
					donotskip = !(skippingtruefrontvalue == skippingscope)
				}
				//adding items to "begincaptures"
				if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
					for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
						val := listitemX.Value.(string)
						lastindex := strings.LastIndex(val, "|")
						if lastindex > -1 {
							scopename := val[0:lastindex]
							scopenumber := val[lastindex+1:]
							middlecapturesmap[scopenumber] = CaptureEntryName{Name: scopename}
						}
					}
				}
				middlearray := make(patternarraytype, 0)
				middlearray = append(middlearray, PatternEntry{Match: middle, Captures: middlecapturesmap})
				//--------------------------------------------------------------------------------------------

				//creating pattern entry
				patternentry := PatternEntry{Begin: begin, End: end, Name: repository.GetScope(listitemwithtype), EndCaptures: endcapturesmap, MorePatterns: middlearray, Comment: realname}
				patternarray = append(patternarray, patternentry)
			} else {

				if !strings.HasPrefix(realname, "_") {
					//fmt.Println("---------------START-------------------") //debug
					//fmt.Println(realname) //debug
                    //getting groups
					groups, _ := getgroups(regex, regex, 0, list.New().Init(), list.New().Init())

					//In the following lines, I am creating the "patterns" field for the json string declared above.
					//the final string will create the json file which will further be converted to plist. In particular,
					// I am creating the items (match and name, alongside the neccessary groups) which will be contained in "patterns" array.

					numberofgroups := groups.Len()
					capturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

					donotskip := true //variable to be used to skip the groups --- kinda a "hack"
					skippingscope := repository.GetScope(listitemwithtype)
					skippingfront := groups.Front()

					if numberofgroups == 1 {
						if skippingfront == nil || skippingscope == defaultscope {
							donotskip = false
						}
						skippingfrontvalue := skippingfront.Value.(string)
						skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
						donotskip = !(skippingtruefrontvalue == skippingscope)
					}

					//fmt.Println(regex)                                   //debug
					//fmt.Println("num groups: ", groups.Len())            //debug
					//fmt.Println("----------------END------------------") //debug
					//adding items to "captures"
					if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
						for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
							val := listitemX.Value.(string)
							//fmt.Println("val is :", val) //debug
							lastindex := strings.LastIndex(val, "|")
							if lastindex > -1 {
								scopename := val[0:lastindex]
								scopenumber := val[lastindex+1:]
								capturesmap[scopenumber] = CaptureEntryName{Name: scopename}
							}
						}
					}
					//creating pattern entry
                    if realname==regex{
                        regex = fmt.Sprintf("(\\A|\\s)%v(\\s)",regex)
                    }
					patternentry := PatternEntry{Match: regex, Name: repository.GetScope(listitemwithtype), Captures: capturesmap, Comment: realname}
					patternarray = append(patternarray, patternentry)
				}
			}
		}

		if *verbose == 1 {
			infolog.Println("Finished processing bnf file.")
		}

		if *doregexorder == 1 {
			//sorting regexes
			infolog.Println("Sorting regexes...")
			sort.Sort(patternarray)
			infolog.Println("Done sorting!")
		}

		if *verbose == 1 {
			infolog.Println("Transforming syntax highlighting data to json...")
		}

		//marshaling output into proper json
		jsonsyntaxobj := JSONSyntax{Name: *name, ScopeName: *scopeName, FileTypes: strings.Split(*fileTypes, ","), Patterns: patternarray, Uuid: u.String()}
		jsonsyntaxobj_result, err := json.MarshalIndent(jsonsyntaxobj, "", "  ")

		if err != nil {
			if *verbose == 1 {
				errlog.Fatalln(fmt.Sprintf("Could not transform syntax highlighting data to json becase %v", err))
			}
		} else {
			if *verbose == 1 {
				infolog.Println("done converting syntax highlighting data to json. Now saving json file...")
			}
			err := ioutil.WriteFile(fmt.Sprintf("%v.tmLanguage.json", *name), jsonsyntaxobj_result, 0644)
			if err != nil {
				errlog.Fatalln(fmt.Sprintf("We a problem writing to file because %v", err))
			} else {
				if *verbose == 1 {
					infolog.Println("json file saved.")
				}
			}
		}

		if *verbose == 1 {
			infolog.Println("Converting json to plist...")
		}
		//convert resulting json to a plist file and save it.
		err = exec.Command("python", "convertor.py", fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v.tmLanguage", *name)).Run()
		if err != nil {
			errlog.Fatalln(fmt.Sprintf("Could not convert json to plist.\n(Reason): %v", err))
		} else {
			if *verbose == 1 {
				infolog.Println("Finished converting json to plist!")
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
		if directoryexists {
			if *verbose == 1 {
				infolog.Println("Found old directory with same name as target directory, deleting...")
			}
			err := os.RemoveAll(*name)
			if err != nil {
				errlog.Fatalln(fmt.Sprintf("Cannot remove old directory because %v", err))
			} else {
				if *verbose == 1 {
					infolog.Println("Old directory with same name as target directory, deleted!")
				}
			}
		}

		//creating folder for syntax highlighting files
		err0 := os.Mkdir(*name, 0775)
		if err0 != nil {
			errlog.Fatalln(fmt.Sprintf("Could not create folder for storing generated files because %v", err0))
		}

		if *verbose == 1 {
			infolog.Println("Moving files into new folder!")
		}
		//moving files into created folder
		err1 := os.Rename(fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v/%v.tmLanguage.json", *name, *name))
		err2 := os.Rename(fmt.Sprintf("%v.tmLanguage", *name), fmt.Sprintf("%v/%v.tmLanguage", *name, *name))
		if err1 != nil || err2 != nil {
			errlog.Fatalln(fmt.Sprintf("Could not move files because %v and/or %v", err1, err2))
		}
	}

	if *verbose == 1 {
		infolog.Println("Finished!")
	}
	os.Exit(0)
}
