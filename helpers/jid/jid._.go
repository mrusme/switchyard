package jid

import (
	"strings"

	"golang.org/x/net/idna"
)

func FromAddress(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return address
	}

	local := address[:at]
	domain := address[at+1:]

	unicode, err := idna.ToUnicode(domain)
	if err != nil {
		return address
	}

	return local + "@" + unicode
}
