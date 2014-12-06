package repository

import (
    "strings"
    "code.google.com/p/gocc/ast"
)
type repoitem struct{
    cleanname string
    realname string
    json string
    
    regexorprod *ast.LexPattern
}

//constructor -- sorta
func NewRepoItem (nameX string) (*repoitem, error){
    
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
    ritem := &repoitem{}
    ritem.cleanname = nameX[startpos:endpos]
    ritem.realname = nameX
    
    return ritem,nil
}
func Getjson (ritem *repoitem) string{
    return ritem.json
}

func Getname (ritem *repoitem) string{
    return ritem.cleanname
}

func GetRealname(ritem *repoitem) string{
    return ritem.realname
}

func SetRighthandside(ritem *repoitem, regexorprodX *ast.LexPattern){
    ritem.regexorprod = regexorprodX
}