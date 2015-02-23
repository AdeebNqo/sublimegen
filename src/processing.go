/*

Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
22 January 2015

*/
package main

import (
	"code.google.com/p/gocc/ast"
	"container/list"
	"fmt"
	"github.com/AdeebNqo/sublimegen/src/repository"
	"strconv"
	"strings"

	"code.google.com/p/gocc/frontend/parser"
	"code.google.com/p/gocc/frontend/scanner"
	"code.google.com/p/gocc/frontend/token"
	"io/ioutil"

	"encoding/json"
	"os"
	"os/exec"

	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"//(documentation: https://godoc.org/github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre)
)

//----------------------------------------------------------------------
//
//   GRAMMAR PARSING
//
//----------------------------------------------------------------------

func getFileAsBuffer(source string) ([]byte, error) {
	srcBuffer, err := ioutil.ReadFile(source)
	return srcBuffer, err
}
func getGrammar(srcBuffer []byte) (*ast.Grammar, error) {
	scanner := &scanner.Scanner{}
	scanner.Init(srcBuffer, token.FRONTENDTokens)
	parser := parser.NewParser(parser.ActionTable, parser.GotoTable, parser.ProductionsTable, token.FRONTENDTokens)
	grammar, err := parser.Parse(scanner)
	return grammar.(*ast.Grammar), err
}
func isSyntaxPartAvailable(grammar *ast.Grammar) bool {
	return isValueNotEmpty(grammar.SyntaxPart)
}
func isValueNotEmpty(grammarPart interface{}) bool {
	return grammarPart != nil
}

//----------------------------------------------------------------------
//
//   PARSING SCOPE VALUES
//
//----------------------------------------------------------------------

func getScopeValues(scopefile string) (map[string]string, error) {
	var data map[string]string
	file, _ := ioutil.ReadFile(*scopesfile)
	err := json.Unmarshal(file, &data)
	return data, err
}

//----------------------------------------------------------------------
//
//   STRING MANIPULATION FUNCTIONS -- INCLUDES CHECKERS.
//
//----------------------------------------------------------------------
func startsAndEndWithQuotation(somestring string) bool {
	return strings.HasPrefix(somestring, "\"") && strings.HasSuffix(somestring, "\"")
}
func removeStartAndEndChars(somestring string) string {
	return somestring[1 : len(somestring)-1]
}

//----------------------------------------------------------------------
//
//   DIRECTORY PROCESSING
//
//----------------------------------------------------------------------
func doesDirExist(name *string) bool {
	directoryexists := true
	if _, err := os.Stat(*name); err != nil {
		if os.IsNotExist(err) {
			directoryexists = false
		}
	}
	return directoryexists
}
func deleteDir(name *string) error {
	err := os.RemoveAll(*name)
	return err
}
func createDir(name *string) error {
	return os.Mkdir(*name, 0775)
}
func moveFiles(name *string) (error, error) {
	jsonerr := os.Rename(fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v/%v.tmLanguage.json", *name, *name))
	langerr := os.Rename(fmt.Sprintf("%v.tmLanguage", *name), fmt.Sprintf("%v/%v.tmLanguage", *name, *name))
	return jsonerr, langerr
}

//----------------------------------------------------------------------
//
//   PLIST - JSON FUNCTIONS
//
//---------------------------------------------------------------------
func convertJSONtoPlist(name *string) error {
	return reallyConvertJSONtoPlist(fmt.Sprintf("%v.tmLanguage.json", *name), fmt.Sprintf("%v.tmLanguage", *name))
}
func reallyConvertJSONtoPlist(jsonfile string, plistfile string) error {
	err := exec.Command("python", "convertor.py", jsonfile, plistfile).Run()
	return err
}

//----------------------------------------------------------------------
//
//   SUBLIMEGEN LOGIC
//
//----------------------------------------------------------------------
func setRegexIfEmpty(regex string, repoitem *repository.Repoitem) {
	if repository.Isregexempty(repoitem) {
		repository.Setregex(repoitem, regex)
	}
}
func verifyRegexIsCorrect(regex string) bool {
	_, compileerr := pcre.Compile(regex, 0)
	if compileerr != nil {
		//regex is not compatile so skip it.
		return false
	}
	return true
}
func addBeginCaptures(groups *list.List, regp pcre.Regexp, donotskip bool, begincapturesmap *map[string]CaptureEntryName) {
	numberofgroups := groups.Len()
	if numberofgroups > 0 && regp.Groups() != 0 && donotskip {
		for listitemX := groups.Front(); listitemX != nil; listitemX = listitemX.Next() {
			val := listitemX.Value.(string)
			lastindex := strings.LastIndex(val, "|")
			if lastindex > -1 {
				scopename := val[0:lastindex]
				scopenumber := val[lastindex+1:]
				(*begincapturesmap)[scopenumber] = CaptureEntryName{Name: scopename}
			}
		}
	}
}
func determineWhetherSkip(groups *list.List, repoitem *repository.Repoitem) bool {
	numberofgroups := groups.Len()
	donotskip := true
	if numberofgroups == 1 {
		skippingscope := repository.GetScope(repoitem)
		skippingfront := groups.Front()
		if skippingfront == nil || skippingscope == defaultscope {
			donotskip = false
		}
		skippingfrontvalue := skippingfront.Value.(string)
		skippingtruefrontvalue := skippingfrontvalue[:strings.LastIndex(skippingfrontvalue, "|")]
		donotskip = !(skippingtruefrontvalue == skippingscope)
	}
	return donotskip
}
func addRepoItemToCollection(repoitem *repository.Repoitem, collection *patternarraytype) patternarraytype {
	realname := repository.GetRealname(repoitem)
	beforealternatives := repository.GetRighthandside(repoitem)
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
	if verifyRegexIsCorrect(regex) {
		//do not add it if it is not.
		return *collection
	}
	regp, _ := pcre.Compile(regex, 0)
	//setting regex
	setRegexIfEmpty(regex, repoitem)
	//determining if one should use begin and end
	usebeginandend, begin, middle, end := determinebeginandend(regex)

	if usebeginandend {
		//sorting out captures for begin regex
		groups, _ := getgroups(begin, begin, 0, list.New().Init(), list.New().Init())
		begincapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

		donotskip := determineWhetherSkip(groups, repoitem)
		//adding items to "begincaptures"
		addBeginCaptures(groups, regp, donotskip, &begincapturesmap)

		//sorting out captures for end regex
		groups, _ = getgroups(end, end, 0, list.New().Init(), list.New().Init())
		endcapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

		donotskip = determineWhetherSkip(groups, repoitem)
		//adding items to "begincaptures"
		addBeginCaptures(groups, regp, donotskip, &endcapturesmap)

		//adding middle parts
		groups, _ = getgroups(middle, middle, 0, list.New().Init(), list.New().Init())
		middlecapturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields
		donotskip = determineWhetherSkip(groups, repoitem)
		//adding items to "begincaptures"
		addBeginCaptures(groups, regp, donotskip, &middlecapturesmap)

		middlearray := make(patternarraytype, 0)
		middlearray = append(middlearray, PatternEntry{Match: middle, Captures: middlecapturesmap})
		//--------------------------------------------------------------------------------------------

		//creating pattern entry
		patternentry := PatternEntry{Begin: begin, End: end, Name: repository.GetScope(repoitem), EndCaptures: endcapturesmap, MorePatterns: middlearray, Comment: realname}
		*collection = append(*collection, patternentry)
	} else {

		if !strings.HasPrefix(realname, "_") {
			//getting groups
			groups, _ := getgroups(regex, regex, 0, list.New().Init(), list.New().Init())

			//In the following lines, I am creating the "patterns" field for the json string declared above.
			//the final string will create the json file which will further be converted to plist. In particular,
			// I am creating the items (match and name, alongside the neccessary groups) which will be contained in "patterns" array.

			capturesmap := make(map[string]CaptureEntryName) //creating map that holds the items of the "captures" fields

			donotskip := determineWhetherSkip(groups, repoitem)
			//adding items to "captures"
			addBeginCaptures(groups, regp, donotskip, &capturesmap)
			//creating pattern entry
			patternentry := PatternEntry{Match: regex, Name: repository.GetScope(repoitem), Captures: capturesmap, Comment: realname}
			*collection = append(*collection, patternentry)
		}
	}
	return *collection
}

//----------------------------------------------------------------------
//
//   REGEX
//
//----------------------------------------------------------------------

/*

Function for disentangling a pattern to obtain it's regex

*/
func constructregexandfillgroups(alternatives []*ast.LexAlt) string {
	regex := ""
	for index, lexitem := range alternatives {
		tmpregex := getregex(lexitem)
		if index > 0 {
			tmpregex = "|" + tmpregex
		}
		regex += tmpregex
	}
	return regex
}

/*

function for retrieving regex from individual lexalt item

return a string, the regex
*/
func getregex(lexitem *ast.LexAlt) string {
	regex := ""
	var tmpoutput string
	var termstring string

	bracedregex := ""
	bracestack := Stack{}
	var strippedtermstring string
	usenormalregex := false

	for _, term := range lexitem.Terms {
		termstring = term.String()

		if strings.HasPrefix(termstring, "'") && strings.HasSuffix(termstring, "'") {
			strippedtermstring = stripliteral(termstring)
		}

		if termstring == "'\\n'" {
			regex += "$"
			bracedregex += "$"
		} else {
			tmpoutput = reallygetregex(term)

			//--------------------------
			//testing if we can close the section
			if strippedtermstring == ")" || strippedtermstring == "}" || strippedtermstring == "]" {
				lastbrace := bracestack.Peek()
				if lastbrace != nil {
					if lastbrace == "(" || lastbrace == "{" || lastbrace == "[" {
						if bracedregex[len(bracedregex)-1] == '(' {
							usenormalregex = true
						}
						bracedregex += ")"
						bracestack.Pop()
					}
				}
				strippedtermstring = ""
			}
			//--------------------------

			regex += tmpoutput
			bracedregex += tmpoutput

			//--------------------------
			//testing if we can open the secion

			//--------------------------
			if strippedtermstring == "[" || strippedtermstring == "(" || strippedtermstring == "{" {
				bracedregex += "("
				bracestack.Push(strippedtermstring)
				strippedtermstring = ""
			}
		}
	}
	if bracestack.Peek() == nil && !usenormalregex {
		return bracedregex
	} else {
		return regex
	}
}

/*

Function for really retrieving regex from individual lexterm items.
This function essentially helps "getregex"

return a string, the regex
*/
func reallygetregex(lexterm interface{}) string {
	switch lexterm.(type) {
	case *ast.LexCharLit:
		{
			term := lexterm.(*ast.LexCharLit)
			termasstring := stripliteral(term.String())

			//The outside character classes are .^$*+?()[{\|
			//The inside character classes are ^-]\

			return Escape(termasstring)
		}
	case *ast.LexCharRange:
		{
			term := lexterm.(*ast.LexCharRange)
			from := reallygetregex(term.From)
			to := reallygetregex(term.To)
			retval := fmt.Sprintf("[%v-%v]", from, to)

			return retval
		}
	case *ast.LexGroupPattern:
		{
			term := lexterm.(*ast.LexGroupPattern)

			retval := "("
			for index, lexalt := range term.LexPattern.Alternatives {
				if index > 0 {
					retval += "|"
				}
				retval += getregex(lexalt)
			}
			retval += ")"
			return retval
		}
	case *ast.LexOptPattern:
		{

			term := lexterm.(*ast.LexOptPattern)

			alternatives := term.LexPattern.Alternatives

			if len(alternatives) == 1 {
				terms := alternatives[0].Terms
				if len(terms) == 1 {
					switch terms[0].(type) {
					case *ast.LexCharLit:
						{
							return fmt.Sprintf("%v?", getregex(alternatives[0]))
						}
					case *ast.LexDot:
						{
							return fmt.Sprintf("%v?", getregex(alternatives[0]))
						}
					}
				}
			}
			retval := "("
			for index, lexalt := range alternatives {
				if index > 0 {
					retval += "|"
				}
				retval += getregex(lexalt)
			}
			retval += ")?"
			return retval
		}
	case *ast.LexRepPattern:
		{
			term := lexterm.(*ast.LexRepPattern)

			alternatives := term.LexPattern.Alternatives

			if len(alternatives) == 1 {
				terms := alternatives[0].Terms
				if len(terms) == 1 {
					switch terms[0].(type) {
					case *ast.LexCharLit:
						{
							return fmt.Sprintf("%v*", getregex(alternatives[0]))
						}
					case *ast.LexDot:
						{
							return fmt.Sprintf("%v*", getregex(alternatives[0]))
						}
					}
				}
			}
			retval := "("
			for index, lexalt := range alternatives {
				if index > 0 {
					retval += "|"
				}
				retval += getregex(lexalt)
			}
			retval += ")*"
			return retval
		}
	case *ast.LexDot:
		{
			return "."
		}
	case *ast.LexRegDefId:
		{
			term := lexterm.(*ast.LexRegDefId)

			for val := repoitems.Front(); val != nil; val = val.Next() {
				rval := val.Value.(*repository.Repoitem)
				if repository.GetRealname(rval) == term.Id {
					if repository.Isregexempty(rval) {

						alternatives := repository.GetRighthandside(rval).Alternatives
						retval := ""
						for index, lexalt := range alternatives {
							if index > 0 {
								retval += "|"
							}
							retval += getregex(lexalt)
						}
						repository.Setregex(rval, retval)
						//return retval
						return fmt.Sprintf("(%s)", retval) //debug

					} else {
						//return repository.Getregex(rval)
						return "(" + repository.Getregex(rval) + ")"
					}
				}
			}
		}
	default:
		{
			return "err"
		}
	}
	return ""
}

/*

Method for escaping special regex characters

*/

func Escape(termasstring string) string {
	if termasstring == "/" {
		return "\\/"
	} else if termasstring == "$" {
		return "\\$"
	} else if termasstring == "*" {
		return "\\*"
	} else if termasstring == "[" {
		return "\\["
	} else if termasstring == "]" {
		return "\\]"
	} else if termasstring == "\\" {
		return "\\\\"
	} else if termasstring == " " {
		return "[ ]"
	} else if termasstring == "+" {
		return "\\+"
	} else if termasstring == "(" {
		return "\\("
	} else if termasstring == ")" {
		return "\\)"
	} else if termasstring == "{" {
		return "\\{"
	} else if termasstring == "}" {
		return "\\}"
	} else if termasstring == "?" {
		return "\\?"
	} else if termasstring == "`" {
		return "\\`"
	} else if termasstring == "." {
		return "\\."
	} else if termasstring == "-" {
		return "-"
	}
	return termasstring
}

/*

Method for strippinf the start and end characters from a token.
It is used mainly to turn 'foobar' to foobar.

*/

func stripliteral(somelit string) string {
	if somelit != "" {
		somelit = somelit[1 : len(somelit)-1]
	}
	return somelit
}

//----------------------------------------------------------------------
//
//   Groups
//
//----------------------------------------------------------------------

/*

This method is used to determine whether the current regex is
is multiline line or not. If it is, it will return true and the begin,
middle and end parts of the regex. This assumes that regex can be seperated
in three. If not, it will return false, and three empty strings.

*/
func determinebeginandend(someregex string) (bool, string, string, string) {
	var begin string
	var end string
	var middle string
	for index, somechar := range someregex {
		//we possibly found the middle part
		if somechar == '(' {
			slice := someregex[index:]
			if len(slice) == 0 {
				return false, "", "", ""
			} else {
				//collecting the middle part
				for innerindex, innersomechar := range slice {
					if innersomechar == ')' {
						middle += ")"

						possibleend := someregex[index+innerindex+1:]
						if len(possibleend) == 0 {
							return false, "", "", ""
						} else {
							if possibleend[0] == '?' || possibleend[0] == '+' || possibleend[0] == '*' {
								middle += string(possibleend[0])
								possibleend = possibleend[1:]
							}
							//collecting end part
							for _, endchar := range possibleend {
								if endchar == '(' || endchar == ')' {
									return false, "", "", ""
								}
								end += string(endchar)
							}
						}
					}
					if len(end) == 0 {
						//else increase middle
						middle += string(innersomechar)
					}
				}
			}
			break
		} else if somechar == '|' {
			return false, "", "", ""
		}
		//appending to the begin char
		begin += string(somechar)
	}
	if len(end) == 0 || len(begin) == 0 {
		return false, "", "", ""
	}
	return true, begin, middle, end
}

/*

This method is for identifiying groups from a regular expressions.
It is will return two lists, The first list contains identified
groups so far, each entry entry will look like the following:

    <somescope>|<number>

Here the somescope variable refers to the scope to be used, e.g constant.numeric
and the number is the number of the group within the regex.

Please note that this method is made to be called kinda-recursively. The second list
the levels are as follows:
is the list of groups in a group level. for instance given the regex (A(B)(C)(D(E))))
Level 1 -  (A(B)(C)(D(E)))
Level 2 - (B), (C), (D(E))
Level 3 - (E)

*/
func getgroups(currentregex string, originalregex string, groupcount int, scopecontainer *list.List, nextlayer *list.List) (*list.List, *list.List) {

	var matched bool
	var scope string

	if currentregex != "" {
		matched, scope = retrievescopefromcapturegroup(currentregex, false)
	}
	if matched {
		groupcount += 1
		scopecontainer.PushBack(scope + "|" + strconv.Itoa(groupcount))
		return scopecontainer, nextlayer
	} else {

		//identifying first opening brace which is not escaped
		bracefound := false
		var tmppos int
		prevregex := ""
		notprevregex := ""
		for braceindex := 0; !bracefound; braceindex-- {
			if braceindex == 0 {
				notprevregex = currentregex
			}

			tmppos = strings.Index(notprevregex, "(")

			if tmppos == -1 {
				//no more braces
				break
			}
			//checking if found brace is not escaped
			escapechars := ""

			for braceindex2 := tmppos + len(prevregex) - 1; braceindex2 >= 0; braceindex2-- {
				if currentregex[braceindex2] == '\\' {
					escapechars += string('\\')
				} else {
					break
				}
			}
			//checking if escaped
			length := len(escapechars)
			if length%2 == 0 {
				//not escaped
				bracefound = true
			} else {
				//the brace has been escaped
				bracefound = false
				prevregex = notprevregex[:tmppos+1]
				notprevregex = notprevregex[tmppos:]
			}
		}
		//found brace position
		pos := tmppos + len(prevregex)

		if pos != -1 {
			//if a brace exists
			regexlength := len(currentregex)
			count := 1 // count of how many braces we have encountered
			stack := &Stack{}

			//identififying group
			stack.Push('(')
			tmpgroup := "("
			for charindex := pos + 1; charindex < regexlength; charindex++ {

				//building group
				if currentregex[charindex] == '(' {
					//checking if opening brace is escaped
					escapechars := ""
					for escapedcheckindex := charindex - 1; escapedcheckindex > 0; escapedcheckindex-- {
						if currentregex[escapedcheckindex] == '\\' {
							escapechars += string('\\')
						} else {
							break
						}
					}
					length := len(escapechars)
					if length%2 == 0 {
						//not escaped
						if count == -1 {
							count = 1
						} else {
							count += 1
						}
					}
				} else if currentregex[charindex] == ')' {
					//checking if opening brace is escaped
					escapechars := ""
					for escapedcheckindex := charindex - 1; escapedcheckindex > 0; escapedcheckindex-- {
						if currentregex[escapedcheckindex] == '\\' {
							escapechars += string('\\')
						} else {
							break
						}
					}
					length := len(escapechars)
					if length%2 == 0 {
						//not escaped
						count -= 1
						if charindex+1 < regexlength && (currentregex[charindex+1] == '*' || currentregex[charindex+1] == '+' || currentregex[charindex+1] == '?') {
							//do nothing
							if count == 0 {
								count = -1
								stack.Push(currentregex[charindex])
								tmpgroup += string(currentregex[charindex])
								continue
							}
						}
					}
				}
				stack.Push(currentregex[charindex])
				tmpgroup += string(currentregex[charindex])

				if count == 0 || count == -1 {
					count = 0

					//seeing if largest group can be matched to a scope selector
					matched, scope = retrievescopefromcapturegroup(tmpgroup, true)
					if matched {
						groupcount += 1
						scopecontainer.PushBack(scope + "|" + strconv.Itoa(groupcount))
					}
					if !matched {
						if strings.HasSuffix(tmpgroup, ")*") {
							//fmt.Println("slice:", tmpgroup[1:len(tmpgroup)-3])
						} else {
							groupcount += 1
							nextlayer.PushFront(tmpgroup[1 : len(tmpgroup)-1])
						}
					}

					if charindex+1 < regexlength {
						if strings.HasPrefix(currentregex[charindex+1:], "|") {
							for it := nextlayer.Back(); it != nil; it = it.Prev() {
								nextlayer.Remove(it)
								itvalue := it.Value.(string)
								scopecontainer, nextlayer = getgroups(itvalue, originalregex, groupcount, scopecontainer, nextlayer)
							}
						}
						//processing reset of regex
						scopecontainer, nextlayer = getgroups(currentregex[charindex+1:], originalregex, groupcount, scopecontainer, nextlayer)
					}
					if nextlayer.Len() != 0 {
						for it := nextlayer.Back(); it != nil; it = it.Prev() {
							nextlayer.Remove(it)
							itvalue := it.Value.(string)
							scopecontainer, nextlayer = getgroups(itvalue, originalregex, groupcount, scopecontainer, nextlayer)
						}
					}
					break
				}
			}
		}
	}
	return scopecontainer, nextlayer
}

func AllEscaped(somestring string) bool {
	stack := &Stack{}
	for _, char := range somestring {
		if char == '\\' {
			stack.Push(char)
		} else {
			curchar := stack.Pop()
			if curchar != '\\' {
				return false
			}
		}
	}
	return true
}

/*

Inefficient method for retrieving scope of regex
*/
func retrievescopefromcapturegroup(capturedregex string, activate bool) (bool, string) {
	if capturedregex != "" {
		if activate == true {
			capturedregex = capturedregex[1 : len(capturedregex)-1]
		}
		for ritem := repoitems.Front(); ritem != nil; ritem = ritem.Next() {

			currreg := repository.Getregex(ritem.Value.(*repository.Repoitem))

			if currreg != "" {
				if currreg == capturedregex {
					tmpscope := repository.GetScope(ritem.Value.(*repository.Repoitem))
					if tmpscope != defaultscope {
						return true, tmpscope
					}
				} else if strings.HasPrefix(capturedregex, "(") && strings.HasSuffix(capturedregex, ")") {
					if currreg == capturedregex[1:len(capturedregex)-1] {
						tmpscope := repository.GetScope(ritem.Value.(*repository.Repoitem))
						if tmpscope != defaultscope {
							return true, tmpscope
						}
					}
				}
			}
		}
	}
	return false, ""
}

/*
Method for determining if strings starts and ends with round braces
*/
func startandendwithrb(somestring string) bool {
	return strings.HasPrefix(somestring, "(") && strings.HasSuffix(somestring, ")")
}
