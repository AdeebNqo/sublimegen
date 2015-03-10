package main

import (
  "testing"
  "io/ioutil"
  )

func TestStartsAndEndWithQuotation(t *testing.T){
	correctString := "\"hello\""
  incorrectStrnig := "hello"
  if (!startsAndEndWithQuotation(correctString)){
      t.Errorf("startsAndEndWithQuotation() does not work for correct strings");
  }
  if (startsAndEndWithQuotation(incorrectStrnig)){
    t.Errorf("startsAndEndWithQuotation() does not work for incorrect strings");
  }
}

func TestDoesDirExist(t *testing.T){
  tempdir,err := ioutil.TempDir(".","")
  if (err!=nil){
    t.Error(err);
  }
  if (!doesDirExist(&tempdir)){
    t.Errorf("doesDirExist does not detect folder that exists.");
  }
}
