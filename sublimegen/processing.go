/*

Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
22 January 2015

*/
package main

import (
	"code.google.com/p/gocc/ast"
	"container/list"
	"fmt"
	"github.com/AdeebNqo/sublimegen/repository"
	"strconv"
	"strings"
)

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
    //fmt.Println("alternatives-->",alternatives,"size:",len(alternatives))
	for index, lexitem := range alternatives {
        tmpregex := getregex(lexitem)
        if index > 0 {
            tmpregex = "|" + tmpregex
        }
        regex += tmpregex
        //fmt.Println("now-->",regex)
	}
	//regex += ")"
	return regex
}

/*

function for retrieving regex from individual lexalt item

return a string, the regex
*/
func getregex(lexitem *ast.LexAlt) string {
	regex := ""
	for _, term := range lexitem.Terms {
        if (term.String()=="'\\n'"){
            regex += "$"
        }else{
		  regex += reallygetregex(term)
        }
	}
	return regex
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
		//fmt.Println("X---X")
		matched, scope = retrievescopefromcapturegroup(currentregex, false)
		//fmt.Println("X---X")
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
			//fmt.Println("current pos:", tmppos+len(prevregex))
		}
		//found brace position
		pos := tmppos + len(prevregex)
		//fmt.Println("pos:", pos)

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
                        if count == -1{
                            count = 1
                        }else{
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
                            if count == 0{
                                count = -1
                            }
						}
					}
				}
				stack.Push(currentregex[charindex])
				tmpgroup += string(currentregex[charindex])

				if count == 0 {

					//fmt.Println("group:", tmpgroup)

					//seeing if largest group can be matched to a scope selector
					matched, scope = retrievescopefromcapturegroup(tmpgroup, true)
					if matched {
						groupcount += 1
						scopecontainer.PushBack(scope + "|" + strconv.Itoa(groupcount))
					}
					if !matched {
						//fmt.Println("inside not matched!")
						if strings.HasSuffix(tmpgroup, ")*") {
							//fmt.Println("slice:", tmpgroup[1:len(tmpgroup)-3])
						} else {
							groupcount += 1
							nextlayer.PushFront(tmpgroup[1:len(tmpgroup)-1])
						}
					}

					if charindex+1 < regexlength {
						//processing reset of regex
						scopecontainer, nextlayer = getgroups(currentregex[charindex+1:], originalregex, groupcount, scopecontainer, nextlayer)
					}

					//fmt.Println("----------New Layer---------")
					//fmt.Println("next layer:")
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

	//fmt.Println("testing match of regex ===> ", capturedregex) //debug
	if capturedregex != "" {
		if activate == true {
			capturedregex = capturedregex[1 : len(capturedregex)-1]
		}
		for ritem := repoitems.Front(); ritem != nil; ritem = ritem.Next() {

			if repository.Getregex(ritem.Value.(*repository.Repoitem)) == capturedregex {
				tmpscope := repository.GetScope(ritem.Value.(*repository.Repoitem))
				if tmpscope != defaultscope {
					//fmt.Println("matched: true, scope:", tmpscope) //debug
					return true, tmpscope
				}
			}
		}
	}
	//fmt.Println("matched: false") //debug
	return false, ""
}

/*
Method for determining if strings starts and ends with round braces
*/
func startandendwithrb(somestring string) bool {
	return strings.HasPrefix(somestring, "(") && strings.HasSuffix(somestring, ")")
}
