/*

These are utility functions
related to the repository items

*/
package repository

import (
	"container/list"
	"fmt"
)

func DoesRepoItemExist(repoitemName string, repoitems *list.List) bool {
	found := false
	for t := repoitems.Front(); t != nil; t = t.Next() {
		item := t.Value.(*Repoitem)
		realname := GetRealname(item)
		if realname == repoitemName {
			found = true
			break
		} else if fmt.Sprintf("\"%v\"", realname) == repoitemName {
			found = true
			break
		}
	}
	return found
}
