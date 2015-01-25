package config

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

/*
   BytesSize represents size written in string format. Examples: "1m", "20g" etc.
   Its main purpose is to be stored and loaded from json.
*/
type BytesSize uint64

// Bytes returns bytes number as uint64
func (b *BytesSize) Bytes() uint64 {
	return uint64(*b)
}

/*
   Parses bytes size such as "1m", "15g" to BytesSize struct.
*/
func BytesSizeFromString(str string) (BytesSize, error) {

	if len(str) < 1 {
		return 0, errors.New("Size string is too small")
	}

	last := strings.ToLower(str[len(str)-1:])

	sizes := map[string]uint64{
		"":  1,
		"k": 1024,
		"m": 1024 * 1024,
		"g": 1024 * 1024 * 1024,
		"t": 1024 * 1024 * 1024 * 1024,
		"z": 1024 * 1024 * 1024 * 1024 * 1024,
	}

	size, ok := sizes[last]
	var num string

	if ok {
		num = str[:len(str)-1]
	} else {
		num = str
		size = 1
	}

	ret, err := strconv.Atoi(num)

	if err != nil {
		return 0, err
	}

	return BytesSize(uint64(ret) * size), nil
}

func (b *BytesSize) UnmarshalJSON(buff []byte) error {
	var buffStr string
	err := json.Unmarshal(buff, &buffStr)
	if err != nil {
		return err
	}
	parsed, err := BytesSizeFromString(buffStr)
	*b = parsed
	return err
}
