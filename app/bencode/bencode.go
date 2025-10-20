package bencode

import (
	"fmt"
	"unicode"
)

type Encoder struct {
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(input string) (any, error) {
	var res any
	if unicode.IsDigit(rune(input[0])) {
		firstColonIndex := -1
		for i, v := range input {
			if v == ':' {
				firstColonIndex = i
				break
			}
		}

		res = input[firstColonIndex+1:]
	} else {
		return "", fmt.Errorf("Only strings are supported at the moment")
	}
	return res, nil
}
