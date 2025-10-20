package bencode

import (
	"fmt"
	"strconv"
	"unicode"
)

type Decoder struct {
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) getType(input string) string {
	if unicode.IsDigit(rune(input[0])) {
		return "string"
	} else if input[0] == 'i' {
		return "integer"
	} else if input[0] == 'l' {
		return "list"
	} else if input[0] == 'd' {
		return "dictionary"
	}
	return ""
}

func (d *Decoder) Decode(input string) (any, error) {
	var res any
	if d.getType(input) == "string" {
		res = d.decodeString(input)
	} else if d.getType(input) == "integer" {
		res = d.decodeInteger(input)
	} else if d.getType(input) == "list" {
		res = d.decodeList(input)
	} else if d.getType(input) == "dictionary" {
		res = d.decodeDictionary(input)
	} else {
		return "", fmt.Errorf("Only strings and integers are supported at the moment")
	}
	return res, nil
}

func (d *Decoder) decodeString(input string) string {
	firstColonIndex := -1
	for i, v := range input {
		if v == ':' {
			firstColonIndex = i
			break
		}
	}
	return input[firstColonIndex+1:]
}

func (d *Decoder) decodeInteger(input string) int {
	length := len(input)
	resStr := input[1 : length-1]
	res, _ := strconv.Atoi(resStr)
	return res
}

func (d *Decoder) decodeList(input string) []any {
	var result []any
	i := 1

	for i < len(input)-1 {
		switch {
		case unicode.IsDigit(rune(input[i])):
			colonIndex := i
			for ; colonIndex < len(input); colonIndex++ {
				if input[colonIndex] == ':' {
					break
				}
			}
			length, _ := strconv.Atoi(input[i:colonIndex])
			start := colonIndex + 1
			end := start + length
			result = append(result, input[start:end])
			i = end

		case input[i] == 'i':
			endIndex := i + 1
			for ; endIndex < len(input); endIndex++ {
				if input[endIndex] == 'e' {
					break
				}
			}
			num, _ := strconv.Atoi(input[i+1 : endIndex])
			result = append(result, num)
			i = endIndex + 1

		case input[i] == 'l':
			subStart := i
			depth := 1
			j := i + 1
			for ; j < len(input); j++ {
				if input[j] == 'l' {
					depth++
				} else if input[j] == 'e' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			sublist := d.decodeList(input[subStart : j+1])
			result = append(result, sublist)
			i = j + 1

		default:
			i++
		}
	}

	return result
}

func (d *Decoder) decodeDictionary(input string) map[string]any {
	res := make(map[string]any)
	i := 1

	for i < len(input) && input[i] != 'e' {
		colonIndex := i
		for ; colonIndex < len(input); colonIndex++ {
			if input[colonIndex] == ':' {
				break
			}
		}
		length, _ := strconv.Atoi(input[i:colonIndex])
		start := colonIndex + 1
		end := start + length
		key := input[start:end]
		i = end
		switch input[i] {
		case 'i':
			endIndex := i + 1
			for ; endIndex < len(input); endIndex++ {
				if input[endIndex] == 'e' {
					break
				}
			}
			num, _ := strconv.Atoi(input[i+1 : endIndex])
			res[key] = num
			i = endIndex + 1

		case 'l':
			startIndex := i
			depth := 1
			j := i + 1
			for ; j < len(input); j++ {
				if input[j] == 'l' {
					depth++
				} else if input[j] == 'e' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			res[key] = d.decodeList(input[startIndex : j+1])
			i = j + 1

		case 'd':
			startIndex := i
			depth := 1
			j := i + 1
			for ; j < len(input); j++ {
				if input[j] == 'd' {
					depth++
				} else if input[j] == 'e' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			res[key] = d.decodeDictionary(input[startIndex : j+1])
			i = j + 1

		default:
			colonIndex := i
			for ; colonIndex < len(input); colonIndex++ {
				if input[colonIndex] == ':' {
					break
				}
			}
			length, _ := strconv.Atoi(input[i:colonIndex])
			start := colonIndex + 1
			end := start + length
			res[key] = input[start:end]
			i = end
		}
	}

	return res
}
