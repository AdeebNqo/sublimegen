package repository

import (
    "strings"
    "code.google.com/p/gocc/ast"
)
type Repoitem struct{
    cleanname string
    realname string
    json string
    scope string
    regex string

    regexorprod *ast.LexPattern
}

//constructor -- sorta
func NewRepoItem (nameX string) (*Repoitem, error){

    //extracting the name of the of the type from the key, remove the "lit"/"var" section
    startpos := 0
    if strings.HasPrefix(nameX,"_") {
        startpos = 1
    }
    endpos := strings.LastIndex(nameX, "_")
    if endpos==-1 || endpos<startpos{
        endpos = len(nameX)
    }
    //creating repository item
    ritem := &Repoitem{}
    ritem.cleanname = nameX[startpos:endpos]
    ritem.realname = nameX
    ritem.regexorprod = nil
    ritem.regex = ""

    return ritem,nil
}
func Getjson (ritem *Repoitem) string{
    return ritem.json
}

func Getname (ritem *Repoitem) string{
    return ritem.cleanname
}

func GetRealname(ritem *Repoitem) string{
    return ritem.realname
}

func SetRighthandside(ritem *Repoitem, regexorprodX *ast.LexPattern){
    ritem.regexorprod = regexorprodX
}

func SetScope(ritem *Repoitem, scope string){
    ritem.scope = scope
}

func Setregex(ritem *Repoitem, regex string){
    ritem.regex = regex
}

func GetDirtyRep(ritem *Repoitem) string{
    return ritem.regexorprod.String()
}

func GetLexpattern(ritem *Repoitem) *ast.LexPattern{
        return ritem.regexorprod
}

func Isregexempty(ritem *Repoitem) (bool){
    return ritem.regex==""
}