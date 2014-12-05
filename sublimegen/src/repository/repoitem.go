package repository

type repoitem struct{
    name string
    json string
}

//constructor -- sorta
func NewRepoItem (nameX string, jsonX string) (*repoitem, error){
    ritem := &repoitem{name:nameX, json:jsonX}
    return ritem,nil
}
func Getjson (ritem *repoitem) string{
    return ritem.json
}

func Getname (ritem *repoitem) string{
    return ritem.name
}