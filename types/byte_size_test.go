package types

import "testing"

func TestByteSizeParsing(t *testing.T) {
	tests := map[string]uint64{
		"500": 500,
		"1m":  1024 * 1024,
		"12k": 12 * 1024,
		"33m": 33 * 1024 * 1024,
		"13g": 13 * 1024 * 1024 * 1024,
	}

	for sizeString, expected := range tests {
		fss, err := BytesSizeFromString(sizeString)
		if err != nil {
			t.Errorf("Error parsing %s: %s", sizeString, err)
		}
		found := fss.Bytes()
		if found != expected {
			t.Errorf("Expected %d for %s but found %d", expected, sizeString, found)
		}
	}

	errors := []string{"1.3g", "lala", "", "1.3l", "1g300m"}

	for _, sizeString := range errors {
		fss, err := BytesSizeFromString(sizeString)
		if err == nil {
			t.Errorf("Expected error for %s but did not get one. Returned %d",
				sizeString, fss)
		}
	}
}
