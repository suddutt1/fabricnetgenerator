package main

import (
	"fmt"
	"strconv"
)

func getMap(element interface{}) map[string]interface{} {
	retMap, ok := element.(map[string]interface{})
	if ok == true {
		return retMap
	}
	return nil
}
func getString(element interface{}) string {
	retString, ok := element.(string)
	if ok == true {
		return retString
	}
	return ""
}
func getNumber(element interface{}) int {

	s := fmt.Sprintf("%v", element)
	retString, err := strconv.Atoi(s)
	if err == nil {
		return retString
	}
	return 0
}
func getBoolean(element interface{}) bool {
	retString, ok := element.(string)
	if ok == true {
		retBool, inValid := strconv.ParseBool(retString)
		if inValid == nil {
			return retBool
		}
	}
	return false
}
func ifExists(mapToCheck map[string]interface{}, attribute string) bool {
	_, ok := mapToCheck[attribute]
	return ok
}
