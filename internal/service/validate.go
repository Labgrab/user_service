package service

import "regexp"

var alphabeticRegexp = regexp.MustCompile("^[\\p{L}\\_\\-\\. ]+$")
var groupCodeRegexp = regexp.MustCompile("^\\p{L}{2,3}\\-[0-9]{1,2}\\-[0-9]{1,2}$")
var phoneNumberRegexp = regexp.MustCompile("^\\+[1-9]\\d{1,14}$")

func ValidateAlphabeticString(userName string) bool {
	return alphabeticRegexp.MatchString(userName)
}

func ValidateGroupCode(groupCode string) bool {
	return groupCodeRegexp.MatchString(groupCode)
}

func ValidatePhoneNumber(phoneNumber string) bool {
	return phoneNumberRegexp.MatchString(phoneNumber)
}
