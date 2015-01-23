/*

Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
22 January 2015

*/
package main
import (
    "code.google.com/p/gocc/ast"
    "fmt"
    "github.com/AdeebNqo/sublimegen/repository"
    "strconv"
    "container/list"
)

//----------------------------------------------------------------------
//
//   REGEX
//
//----------------------------------------------------------------------


/*

Function for disentangling a pattern to obtain it's regex

*/
func constructregexandfillgroups(alternatives []*ast.LexAlt) string{
    regex := ""
    for index,lexitem := range alternatives{
        tmpregex := getregex(lexitem)
        if index>0{
            tmpregex = " | " + tmpregex
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
    for _,term := range lexitem.Terms{
        regex += reallygetregex(term)
    }
    return regex
}
/*

Function for really retrieving regex from individual lexterm items.
This function essentially helps "getregex"

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
            }else if termasstring=="+" {
                return "\\+"
            }else if termasstring=="("{
                return "\\("
            }else if termasstring==")"{
                return "\\)"
            }else if termasstring=="{"{
                return "\\{"
            }else if termasstring=="}"{
                return "\\}"
            }
            
            return termasstring
        }
        case *ast.LexCharRange:{
            term := lexterm.(*ast.LexCharRange)
            from := reallygetregex(term.From)
            to := reallygetregex(term.To)
            retval := fmt.Sprintf("[%v-%v]", from, to)
            
            return retval
        }
        case *ast.LexGroupPattern:{
            term := lexterm.(*ast.LexGroupPattern)
            
            retval := "("
            for index,lexalt := range term.LexPattern.Alternatives{
                if index>0{
                    retval += "|"
                }
                retval += getregex(lexalt)
            }
            retval += ")"
            return retval
        }
        case *ast.LexOptPattern:{
            
            term := lexterm.(*ast.LexOptPattern)

            alternatives := term.LexPattern.Alternatives

            if len(alternatives)==1{
                terms := alternatives[0].Terms
                if len(terms)==1{
                    switch terms[0].(type){
                        case *ast.LexCharLit:{
                            return fmt.Sprintf("%v?", getregex(alternatives[0]))
                        }
                        case *ast.LexDot:{
                            return fmt.Sprintf("%v?", getregex(alternatives[0]))
                        }
                    }
                }
            }
            retval := "("
            for index,lexalt := range alternatives{
                if index>0{
                    retval += "|"
                }
                retval += getregex(lexalt)
            }
            retval += ")?"
            return retval
        }
        case *ast.LexRepPattern:{
            term := lexterm.(*ast.LexRepPattern)
            
            alternatives := term.LexPattern.Alternatives
            
            if len(alternatives)==1{
                terms := alternatives[0].Terms
                if len(terms)==1{
                    switch terms[0].(type){
                        case *ast.LexCharLit:{
                            return fmt.Sprintf("%v*", getregex(alternatives[0]))
                        }
                        case *ast.LexDot:{
                            return fmt.Sprintf("%v*", getregex(alternatives[0]))
                        }
                    }
                }
             }
            retval := "("
            for index,lexalt := range alternatives{
                if index>0{
                    retval += "|"
                }
                retval += getregex(lexalt)
            }
            retval += ")*"
            return retval
        }
        case *ast.LexDot:{
            return "."
        }
        case *ast.LexRegDefId:{
            term := lexterm.(*ast.LexRegDefId)
            
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
                        //return fmt.Sprintf("(%s)",retval) //debug
                        
                    }else{
                        return "("+repository.Getregex(rval)+")"
                    }
                }
            }
        }
        default:{
            return "err"
        }
    }
    return ""
}

/*

Method for strippinf the start and end characters from a token.
It is used mainly to turn 'foobar' to foobar.

*/

func stripliteral(somelit string) string{
    if somelit != "" {
        somelit = somelit[1:len(somelit)-1]
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
func determinebeginandend(someregex string) (bool, string, string, string){
    var begin string
    var end string
    var middle string
    for index, somechar := range someregex{
        //we possibly found the middle part
        if somechar=='('{
            slice := someregex[index:]
            if len(slice)==0{
                return false, "", "", ""
            }else{
                //collecting the middle part
                for innerindex,innersomechar := range slice{
                    if innersomechar==')'{
                        middle += ")"
                        
                        possibleend := someregex[index+innerindex+1:]
                        if len(possibleend)==0{
                            return false,"","",""
                        }else{
                            if possibleend[0]=='?' || possibleend[0]=='+' || possibleend[0]=='*'{
                                middle += string(possibleend[0])
                                possibleend = possibleend[1:]
                            }
                            //collecting end part
                            for _, endchar := range possibleend{
                                if endchar=='(' || endchar==')'{
                                    return false,"","",""
                                }
                                end += string(endchar)
                            }
                        }
                    }
                    if len(end)==0{
                        //else increase middle
                        middle += string(innersomechar)
                    }
                }
            }
            break
        }
        //appending to the begin char
        begin += string(somechar)
    }
    if len(end)==0 || len(begin)==0{
        return false,"","",""
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
Level 1 -  (A(B)(C)(D(E))))
Level 2 - (B), (C), (D(E))
Level 3 - (E)

*/
func getgroups(currentregex string, originalregex string,groupcount int, scopecontainer *list.List, nextlayer *list.List) (*list.List, *list.List){
    
    var matched bool
    var scope string
    
    //if currentregex==originalregex{
    matched, scope = retrievescopefromcapturegroup(currentregex, false)
    //}
    if matched{
        
        groupcount+=1
        scopecontainer.PushBack(scope+"|"+strconv.Itoa(groupcount))
        return scopecontainer, nextlayer
    }else{
        
        count := 0
        regexlength := len(currentregex)
        stack := Stack{}

        STARTBRACE:
            for pos:=0; pos<regexlength; pos++{
                currcharacter := string(currentregex[pos])
                if currcharacter=="(" {
                    if pos==0 || string(currentregex[pos-1])!="\\"{
                        
                        //found starting brace
                        count+=1
                        stack.Push(currcharacter)
                        //pushing in the rest till we find matching brace
                        for innerindex:=pos+1; innerindex<regexlength; innerindex++{

                            currcharacter2 := string(currentregex[innerindex])
                            if currcharacter2==")"{
                                if innerindex!=0 || string(currentregex[innerindex-1])!="\\"{
                                    count-=1
                                    stack.Push(currcharacter2)

                                    //checking if following character is a *, + or ?
                                    if innerindex+1 < len(currentregex){
                                        charafterrbrace := string(currentregex[innerindex+1])
                                        if charafterrbrace=="*" || charafterrbrace=="+" || charafterrbrace=="?"{
                                            innerindex+=1;
                                            stack.Push(charafterrbrace)
                                        }
                                    }
                                }
                            }else if currcharacter2=="("{
                                if innerindex==0 || string(currentregex[innerindex-1])!="\\"{
                                    count+=1
                                    stack.Push(currcharacter2)
                                }
                            }else{
                                stack.Push(currcharacter2)
                            }
                            if count==0{
                                biggest := ""
                                for {
                                    head := stack.Pop()
                                    if head==nil{
                                        break
                                    }
                                    biggest = head.(string)+biggest
                                }
                                //TODO: Check if this matches any special definitions.
                                // if it doesnt, and there is still other braces after the  the closing one -- process the following ones.
                                // else process the items within the braces. ||
                                //                                           \/
                                //fmt.Println(biggest, groupcount)
                                matched, scope := retrievescopefromcapturegroup(biggest, true)
                                if matched{
                                    groupcount+=1
                                    scopecontainer.PushBack(scope+"|"+strconv.Itoa(groupcount))
                                }
                                if !matched{
                                    groupcount +=1
                                    //scopecontainer = getgroups(biggest[1:len(biggest)-1], originalregex ,groupcount, alternatives, scopecontainer)
                                    nextlayer.PushFront(biggest)
                                }
                                if (len(biggest) < regexlength) {
                                    scopecontainer,nextlayer = getgroups(currentregex[innerindex+1:], originalregex, groupcount, scopecontainer, nextlayer)
                                }
                                if nextlayer.Len()!=0{
                                    for it := nextlayer.Back(); it !=nil; it = it.Prev(){
                                        nextlayer.Remove(it)
                                        itvalue := it.Value.(string)
                                        scopecontainer,nextlayer = getgroups(itvalue[1:len(itvalue)-1], originalregex, groupcount, scopecontainer, nextlayer)
                                    }
                                }
                                break STARTBRACE
                            }
                        }
                        break
                    }
                }
            }
    }
    return scopecontainer, nextlayer
}

/*

Inefficient method for retrieving scope of regex
*/
func retrievescopefromcapturegroup(capturedregex string, activate bool) (bool, string){
    if activate{
        capturedregex = capturedregex[1:len(capturedregex)-1]
    }
    for ritem := repoitems.Front(); ritem !=nil; ritem = ritem.Next(){
        if repository.Getregex(ritem.Value.(*repository.Repoitem)) == capturedregex{
            tmpscope := repository.GetScope(ritem.Value.(*repository.Repoitem))
            //fmt.Println("name:",repository.GetRealname(ritem.Value.(*repository.Repoitem)), ",scope:",tmpscope)
            if tmpscope!=defaultscope{
                //fmt.Println("matched!")
                return true,tmpscope
            }
        }
    }
    return false, ""
}