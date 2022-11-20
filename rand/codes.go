package rand

import "math/rand"

const BaseCharsCount = 62

// Generates random alphanumeric codes of given length.
func AlphanumStr(length int) string {
	code := make([]byte, length)
	chars := activationCodeChars()

	for i := 0; i < length; i++ {
		code[i] = chars[rand.Intn(BaseCharsCount)]
	}

	return string(code)
}

func activationCodeChars() [BaseCharsCount]byte {
	return [...]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',

		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',

		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
		'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	}
}
