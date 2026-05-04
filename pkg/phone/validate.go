package phone

import "regexp"

var rePhone = regexp.MustCompile(`^\d{11}$`)

func IsValid(s string) bool { return rePhone.MatchString(s) }
