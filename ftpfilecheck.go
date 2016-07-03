// Copyright © 2016 Jörg Kost, joerg.kost@gmx.com
// ftpfilecheck-plugin for nagios or icinga
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

import (
	"flag"
	"fmt"
	"github.com/jlaffaye/ftp"
	"os"
	"time"
)

const (
	stateOk        = 0 // Will signal OK and exit 0 to Nagios / Icinga
	stateWarning   = 1 // Will signal WARNING and exit 1 to Nagios / Icinga
	stateFail      = 2 // Will signal CRITICAL and exit 2 to Nagios / Icinga
	ftpCmdOk       = "Found %s, size is %d"
	ftpCantConnect = "Cant connect to server"
	ftpCantFind    = "Cant find %s"
	ftpCantLogin   = "Authentication failed"
	ftpCantCmd     = "Cant send command to server"
	ftpCantList    = "Directory listing not available"
	ftpFileWrong   = "%s has wrong size or is empty, size is %d"
	MaxUint        = ^uint64(0)
)

var nagios = map[int]string{
	stateOk:      "OK",
	stateWarning: "WARNING",
	stateFail:    "CRITICAL",
}

var hostPort = flag.String("hostPort", "ip:21", "ip and port of the ftp-server, e.g. ftp.example.com:21")
var login = flag.String("login", "MyUsername", "ftp-login")
var password = flag.String("password", "MyPassword", "ftp-password")
var logDir = flag.String("logdir", "/log/", "sub-directory for our wanted file")
var fileName = flag.String("filename", "access_log", "filename we are looking for")
var fileDelim = flag.String("delim", "-", "adds given delimeter between fileName and currentDate if addToday was set")
var fileSuffix = flag.String("suffix", "", "possible suffix that will be added to the filename")
var addToday = flag.Bool("date", false, "adds suffix of the current date in form %YY-%MM-%DD to the filename")
var addYesterday = flag.Bool("yesterday", false, "add suffix of yesterday in form %YY-%MM-%DD to the filename")
var minSize = flag.Uint64("minsize", 1, "minimum shall be 1 byte ")
var maxSize = flag.Uint64("maxsize", MaxUint, "maximum  Uint64 size")

func main() {
	var ftpStatus = stateFail
	var ftpMessage = "Unbekannt"
	var FilenameFull string
	var FileFound = false
	var files []*ftp.Entry

	flag.Parse()

	t := time.Now()
	ty := t.Add(-24 * time.Hour)

	conn, err := ftp.Dial(*hostPort)
	if err != nil {
		ftpMessage = ftpCantConnect
		fmt.Println(*hostPort)
		goto printError
	}

	err = conn.Login(*login, *password)
	if err != nil {
		ftpMessage = ftpCantLogin
		goto printError
	}

	err = conn.NoOp()
	if err != nil {
		ftpMessage = ftpCantCmd
		goto printError
	}

	if *addToday == true {
		FilenameFull = fmt.Sprintf("%s%s%02d-%02d-%02d%s", *fileName, *fileDelim, t.Year(), t.Month(), t.Day(),
			*fileSuffix)
	} else if *addYesterday == true {
		FilenameFull = fmt.Sprintf("%s%s%02d-%02d-%02d%s", *fileName, *fileDelim, ty.Year(), ty.Month(), ty.Day(),
			*fileSuffix)
	} else {
		FilenameFull = fmt.Sprintf("%s%s", *fileName, *fileSuffix)
	}

	files, err = conn.List(*logDir)

	for _, v := range files {
		if v.Name == FilenameFull {
			FileFound = true
			if uint64(v.Size) <= *minSize {
				ftpMessage = fmt.Sprintf(ftpFileWrong, FilenameFull, v.Size)
				ftpStatus = stateWarning
			} else if uint64(v.Size) > *maxSize {
				ftpMessage = fmt.Sprintf(ftpFileWrong, FilenameFull, v.Size)
				ftpStatus = stateWarning
			} else {
				ftpMessage = fmt.Sprintf(ftpCmdOk, FilenameFull, v.Size)
				ftpStatus = stateOk
			}
		}
	}

	if err != nil {
		ftpMessage = ftpCantList
		ftpStatus = stateFail
	}

	if FileFound == false {
		ftpMessage = fmt.Sprintf(ftpCantFind, FilenameFull)
		ftpStatus = stateFail
	}

	conn.Quit()

printError:
	fmt.Printf("%s %s\n", nagios[ftpStatus], ftpMessage)
	os.Exit(ftpStatus)
}
