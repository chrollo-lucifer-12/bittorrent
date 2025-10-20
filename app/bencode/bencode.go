package bencode

import (
	"fmt"
	"strconv"
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
	} else if input[0] == 'i' {
		length := len(input)
		resStr := input[1 : length-1]
		res, _ = strconv.Atoi(resStr)
	} else {
		return "", fmt.Errorf("Only strings and integers are supported at the moment")
	}
	return res, nil
}
