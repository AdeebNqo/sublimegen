/*

Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
22 January 2014

*/
package logger

import (
        "log"
        "io"
        )

type logger struct{
    Trace *log.Logger
    Info *log.Logger
    Warning *log.Logger
    Error *log.Logger
}

func Init(
    infoHandle io.Writer,
    warningHandle io.Writer,
    errorHandle io.Writer) logger{

    somelogger := logger{}
    somelogger.Info = log.New(infoHandle,
        "INFO: ",
        log.Ldate|log.Ltime)

    somelogger.Warning = log.New(warningHandle,
        "WARNING: ",
        log.Ldate|log.Ltime)

    somelogger.Error = log.New(errorHandle,
        "ERROR: ",
        log.Ldate|log.Ltime)
    return somelogger
}

func (somelogger *logger) Warn(warning string){
    somelogger.Warning.Println(warning)
}
func (somelogger *logger) Inform(info string){
    somelogger.Info.Println(info)
}
func (somelogger *logger) Err(err string){
    somelogger.Error.Println(err)
}

