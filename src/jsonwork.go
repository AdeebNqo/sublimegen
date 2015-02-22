/*

Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
22 January 2015

*/
package main

import (
	"os/exec"
	"strings"
)

//the following structs are used to produce a json file
type JSONSyntax struct {
	Name      string         `json:"name"`
	ScopeName string         `json:"scopeName"`
	FileTypes []string       `json:"fileTypes"`
	Patterns  []PatternEntry `json:"patterns,omitempty"`
	Uuid      string         `json:"uuid"`
}
type PatternEntry struct {
	Match string `json:"match,omitempty"`

	Begin         string                      `json:"begin,omitempty"`
	BeginCaptures map[string]CaptureEntryName `json:"beginCaptures,omitempty"`

	End         string                      `json:"end,omitempty"`
	EndCaptures map[string]CaptureEntryName `json:"endCaptures,omitempty"`

	Name     string                      `json:"name,omitempty"`
	Captures map[string]CaptureEntryName `json:"captures,omitempty"`

	MorePatterns []PatternEntry `json:"patterns,omitempty"`

	Comment string `json:"_comment,omitempty"`
}
type CaptureEntryName struct {
	Name string `json:"name,omitempty"`
}

//implenting sort interface in order to sort Patterns []PatternEntry from JSONSyntax
//by length of regex
type patternarraytype []PatternEntry

func (p patternarraytype) Len() int {
	return len(p)
}
func (p patternarraytype) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p patternarraytype) Less(i, j int) bool {

	if p[i].Match == "" || p[j].Match == "" {

		if len(p[i].Match) != len(p[j].Match) {
			return len(p[i].Match) < len(p[j].Match)
		}
		return p[i].Match < p[j].Match
	}

	pythonexecutable := "python" ///tmp/pypy-2.4.0-linux64/bin/pypy"

	cmd := exec.Command(pythonexecutable, "greenery/compare.py", p[i].Match, p[j].Match)
	output, err := cmd.CombinedOutput()
	outputString := string(output)

	if outputString == "subset" {
		return true
	}

	cmd2 := exec.Command(pythonexecutable, "greenery/compare.py", p[j].Match, p[i].Match)
	output2, err2 := cmd2.CombinedOutput()
	output2String := string(output2)

	if err == nil && err2 == nil {
		outputString = strings.TrimSpace(outputString)
		output2String = strings.TrimSpace(output2String)
		if outputString == "notsubset" && output2String == "notsubset" {
			return p[i].Match < p[j].Match
		} else if outputString == "subset" {
			return true
		} else if outputString == "notsubset" {
			return false
		}
	}
	if len(p[i].Match) != len(p[j].Match) {
		return len(p[i].Match) < len(p[j].Match)
	}
	return p[i].Match < p[j].Match
}
