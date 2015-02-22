/*

Copyright 2014 Zola Mahlaza <adeebnqo@gmail.com>
02 December 2014

This is the driver class. It is resposible for creating the
BNF loader/parser and generating the sublime text syntax highligting
file.

*/

package main

import (
	"io/ioutil"

	"container/list"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/AdeebNqo/sublimegen/src/repository"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre" //(documentation: https://godoc.org/github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre)
	"github.com/nu7hatch/gouuid"
	"log"
	"os"
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

var scopeReferences map[string]string //used to map token to scopes

func main() {
	flag.Parse() //parsing commandline flags
	//initializing logging objects
	errlog = log.New(os.Stdout, "Error: ", log.Ltime|log.Lshortfile)
	infolog = log.New(os.Stderr, "Info: ", log.Ltime|log.Lshortfile)
	//initializing value of the default scope
	defaultscope = fmt.Sprintf("source.%v", *fileTypes) //TODO: fileTypes is a comma seperated list
	// thefore this is wrong. split fileTypes using
	// a command and probably only use the first value
	// in the resulting array.
	//reading source file
	srcBuffer, err := getFileAsBuffer(*source)
	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Cannot read file because %v", err))
	}
	//parsing input and obtaining the grammar
	grammar, err := getGrammar(srcBuffer)
	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Parse error: %v", err))
	}
	//loading tokens and scopes
	scopeReferences, err = getScopeValues(*scopesfile)
	if err != nil {
		errlog.Fatalln(fmt.Sprintf("Cannot parse json file with scopes because %v", err))
	}
	//instantiating the list that will contain all the repoitems, that is, pattern field entries.
	repoitems = list.New()
	repoitems.Init()
	/*

	   Processing productions

	*/
	if isValueNotEmpty(grammar.LexPart) && isValueNotEmpty(grammar.LexPart.ProdList) && isValueNotEmpty(grammar.LexPart.ProdList.Productions) {
		productions := grammar.LexPart.ProdList.Productions
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
			somescope := scopeReferences[repository.GetRealname(patternobj)]
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
	if isSyntaxPartAvailable(grammar) && isValueNotEmpty(grammar.SyntaxPart.ProdList) {
		for _, synprod := range grammar.SyntaxPart.ProdList {
			for _, synsymb := range synprod.Body.Symbols {
				synprodname := synsymb.String()
				found := repository.DoesRepoItemExist(synprodname, repoitems)
				if !found {
					if startsAndEndWithQuotation(synprodname) {
						//creating new repoitem
						productionId := removeStartAndEndChars(synprodname)
						patternobj, err := repository.NewRepoItem(productionId)
						if err != nil {
							//ignoring token -- since there was a problem creating it.
							errlog.Fatalln(fmt.Sprintf("could not process %v. reason: %v", productionId, err))
							break
						}
						repository.SetRighthandside(patternobj, nil)
						//getting and setting the scope.
						somescope := scopeReferences[repository.GetRealname(patternobj)]
						if somescope != "" {
							repository.SetScope(patternobj, somescope)
						} else {
							repository.SetScope(patternobj, defaultscope)
						}
						//ignoring repoitems whose id does not contain letters
						//and those which have the default scope.
						if strings.ContainsAny(productionId, "abcdefghijklmnopqrstuvwxyz") && somescope != defaultscope {
							repoitems.PushBack(patternobj)
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
		patternarray := make(patternarraytype, 0)
		//0. Generate repository field from bnf file
		for listitem := repoitems.Front(); listitem != nil; listitem = listitem.Next() {
			listitemwithtype := listitem.Value.(*repository.Repoitem)
			realname := repository.GetRealname(listitemwithtype)
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
			if realname == regex || realname[1:] == regex {
				regex = fmt.Sprintf("(\\b)%v(\\b)", regex)
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
					//adding items to "captures"
					if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
						for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
							val := listitemX.Value.(string)
							lastindex := strings.LastIndex(val, "|")
							if lastindex > -1 {
								scopename := val[0:lastindex]
								scopenumber := val[lastindex+1:]
								capturesmap[scopenumber] = CaptureEntryName{Name: scopename}
							}
						}
					}
					//creating pattern entry
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
			if *verbose == 1 {
				infolog.Println("Sorting regexes...")
			}
			sort.Sort(patternarray)
			if *verbose == 1 {
				infolog.Println("Done sorting!")
			}
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
		err = convertJSONtoPlist(name)
		if err != nil {
			errlog.Fatalln(fmt.Sprintf("Could not convert json to plist.\n(Reason): %v", err))
		} else {
			if *verbose == 1 {
				infolog.Println("Finished converting json to plist!")
			}
		}
		//removing old directory with the same name
		if doesDirExist(name) {
			if *verbose == 1 {
				infolog.Println("Found old directory with same name as target directory, deleting...")
			}
			err := deleteDir(name)
			if err != nil {
				errlog.Fatalln(fmt.Sprintf("Cannot remove old directory because %v", err))
			} else {
				if *verbose == 1 {
					infolog.Println("Old directory with same name as target directory, deleted!")
				}
			}
		}

		//creating folder for syntax highlighting files
		err0 := createDir(name)
		if err0 != nil {
			errlog.Fatalln(fmt.Sprintf("Could not create folder for storing generated files because %v", err0))
		}
		if *verbose == 1 {
			infolog.Println("Moving files into new folder!")
		}
		//moving files into created folder
		err1, err2 := moveFiles(name)
		if err1 != nil || err2 != nil {
			errlog.Fatalln(fmt.Sprintf("Could not move files because %v and/or %v", err1, err2))
		}
	}
	if *verbose == 1 {
		infolog.Println("Finished!")
	}
	os.Exit(0)
}
