package bencode

import (
	"fmt"
	"sort"
	"strconv"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(input []byte) (any, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	switch input[0] {
	case 'i':
		val, _, err := decodeInteger(input, 0)
		return val, err
	case 'l':
		val, _, err := decodeList(input, 0)
		return val, err
	case 'd':
		val, _, err := decodeDictionary(input, 0)
		return val, err
	default:
		val, _, err := decodeString(input, 0)
		return val, err
	}
}

func decodeString(input []byte, start int) ([]byte, int, error) {
	colon := start
	for ; colon < len(input); colon++ {
		if input[colon] == ':' {
			break
		}
	}
	if colon == len(input) {
		return nil, 0, fmt.Errorf("colon not found in string field")
	}

	length, err := strconv.Atoi(string(input[start:colon]))
	if err != nil {
		return nil, 0, err
	}

	dataStart := colon + 1
	dataEnd := dataStart + length
	if dataEnd > len(input) {
		return nil, 0, fmt.Errorf("unexpected end of input for string")
	}

	return input[dataStart:dataEnd], dataEnd, nil
}

func decodeInteger(input []byte, start int) (int, int, error) {
	if input[start] != 'i' {
		return 0, 0, fmt.Errorf("expected 'i' at start of integer")
	}
	end := start + 1
	for ; end < len(input); end++ {
		if input[end] == 'e' {
			break
		}
	}
	if end == len(input) {
		return 0, 0, fmt.Errorf("unterminated integer")
	}

	val, err := strconv.Atoi(string(input[start+1 : end]))
	if err != nil {
		return 0, 0, err
	}
	return val, end + 1, nil
}

func decodeList(input []byte, start int) ([]any, int, error) {
	if input[start] != 'l' {
		return nil, 0, fmt.Errorf("expected 'l' at start of list")
	}

	var res []any
	i := start + 1
	for i < len(input) && input[i] != 'e' {
		switch input[i] {
		case 'i':
			val, next, err := decodeInteger(input, i)
			if err != nil {
				return nil, 0, err
			}
			res = append(res, val)
			i = next
		case 'l':
			val, next, err := decodeList(input, i)
			if err != nil {
				return nil, 0, err
			}
			res = append(res, val)
			i = next
		case 'd':
			val, next, err := decodeDictionary(input, i)
			if err != nil {
				return nil, 0, err
			}
			res = append(res, val)
			i = next
		default:
			val, next, err := decodeString(input, i)
			if err != nil {
				return nil, 0, err
			}
			res = append(res, val) // keep as []byte
			i = next
		}
	}

	if i >= len(input) {
		return nil, 0, fmt.Errorf("unterminated list")
	}

	return res, i + 1, nil
}

func decodeDictionary(input []byte, start int) (map[string]any, int, error) {
	if input[start] != 'd' {
		return nil, 0, fmt.Errorf("expected 'd' at start of dictionary")
	}

	res := make(map[string]any)
	i := start + 1
	for i < len(input) && input[i] != 'e' {
		keyBytes, next, err := decodeString(input, i)
		if err != nil {
			return nil, 0, err
		}
		key := string(keyBytes) // keys are always strings in Bencode
		i = next

		if i >= len(input) {
			return nil, 0, fmt.Errorf("unexpected end of dictionary")
		}

		switch input[i] {
		case 'i':
			val, next, err := decodeInteger(input, i)
			if err != nil {
				return nil, 0, err
			}
			res[key] = val
			i = next
		case 'l':
			val, next, err := decodeList(input, i)
			if err != nil {
				return nil, 0, err
			}
			res[key] = val
			i = next
		case 'd':
			val, next, err := decodeDictionary(input, i)
			if err != nil {
				return nil, 0, err
			}
			res[key] = val
			i = next
		default:
			val, next, err := decodeString(input, i)
			if err != nil {
				return nil, 0, err
			}
			res[key] = val // keep as []byte
			i = next
		}
	}

	if i >= len(input) {
		return nil, 0, fmt.Errorf("unterminated dictionary")
	}

	return res, i + 1, nil
}

type Encoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(input any) ([]byte, error) {
	switch v := input.(type) {
	case int:
		return []byte(e.encodeInt(v)), nil
	case string:
		return []byte(e.encodeString([]byte(v))), nil
	case []byte:
		return []byte(e.encodeString(v)), nil
	case []any:
		return e.encodeList(v)
	case map[string]any:
		return e.encodeDict(v)
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func (e *Encoder) encodeInt(i int) string {
	return fmt.Sprintf("i%de", i)
}

func (e *Encoder) encodeString(b []byte) string {
	return fmt.Sprintf("%d:%s", len(b), string(b))
}

func (e *Encoder) encodeList(list []any) ([]byte, error) {
	result := []byte("l")
	for _, item := range list {
		encoded, err := e.Encode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, encoded...)
	}
	result = append(result, 'e')
	return result, nil
}

func (e *Encoder) encodeDict(dict map[string]any) ([]byte, error) {
	result := []byte("d")
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result = append(result, []byte(e.encodeString([]byte(k)))...)
		encodedValue, err := e.Encode(dict[k])
		if err != nil {
			return nil, err
		}
		result = append(result, encodedValue...)
	}
	result = append(result, 'e')
	return result, nil
}
