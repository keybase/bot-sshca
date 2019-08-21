package kssh

import (
	"fmt"
	"strings"
)

func DebugLog(runtimeConfig RuntimeConfig, fmtString string, a ...interface{}) {
	if runtimeConfig.Debug {
		str := "kssh: " + fmt.Sprintf(fmtString, a...)
		if !strings.HasSuffix(str, "\n") {
			str += "\n"
		}
		fmt.Print(str)
	}
}
