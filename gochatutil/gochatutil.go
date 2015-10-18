package gochatutil

import (
	"log"
	"os"
)

type Err_code int

const (
	_            Err_code = iota
	ERR_CONTINUE          = iota
	ERR_STOP              = iota
)

func CheckError(err error) bool {
	return CheckErrorAndAct(err, ERR_CONTINUE)
}

func CheckErrorAndAct(err error, action Err_code) bool {
	if err != nil {
		log.Fatal("Error %v\n", err)
		switch action {
		case ERR_STOP:
			os.Exit(-1)
		case ERR_CONTINUE:
			return true
		default:
			return true
		}
	}
	return false
}
