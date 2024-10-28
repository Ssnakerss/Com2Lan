package tools

import "log"

func Fail(err error, message string) {
	if err != nil {
		log.Fatal(message, "err:", err)
	}
}
