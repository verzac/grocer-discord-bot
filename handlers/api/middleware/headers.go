package middleware

import (
	"errors"
	"strings"
)

const (
	// mostly used by API key users
	HeaderTypeBasic = "Basic"
	// used by OAuth2 Discord users
	HeaderTypeBearer = "Bearer"
)

var (
	allowedHeaderTypesMap = map[string]bool{
		HeaderTypeBasic:  true,
		HeaderTypeBearer: true,
	}
	ErrMalformedAuthorizationHeader = errors.New("malformed authorization header")
)

func getHeaderTypeAndValue(header string) (headerType string, headerValue string, err error) {
	if header == "" {
		return "", "", errors.New("missing authorization header")
	}
	headerTokens := strings.Split(header, " ")
	if len(headerTokens) != 2 {
		return "", "", ErrMalformedAuthorizationHeader
	}
	headerType = headerTokens[0]
	headerValue = headerTokens[1]
	if allowed, _ := allowedHeaderTypesMap[headerType]; !allowed {
		return "", "", ErrMalformedAuthorizationHeader
	}
	return
}
