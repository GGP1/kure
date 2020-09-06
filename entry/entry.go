package entry

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
)

// Password characters.
var (
	def       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"                                                            // default
	lowerCase = "abcdefghijklmnopqrstuvwxyz"                                                                                                // level 1
	upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"                                                                                                // level 2
	digits    = "0123456789"                                                                                                                // level 3
	space     = " "                                                                                                                         // level 4
	brackets  = "(){}[]<>"                                                                                                                  // level 5
	points    = "_.¿?!¡,;:"                                                                                                                 // level 6
	special   = "$%&|/=*#@=~€^"                                                                                                             // level 7
	extended  = "€ƒ„…†‡0ˆ‰Š‹›ŒŽ‘’“”•-+_—˜™šœžŸ¢£¤¥¦§¨©ª«¬®¯°±²³´µ¶·¸¹º»¼½¾ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõö÷øùúûüýþÿ" // level 8
)

// New creates a new entry.
func New(title, username, password, url, expires string, secure bool) *Entry {
	return &Entry{
		Title:    []byte(title),
		Username: []byte(username),
		Password: []byte(password),
		URL:      []byte(url),
		Expires:  []byte(expires),
		Secure:   secure,
	}
}

// GeneratePassword generates a random password with the length and format given.
func GeneratePassword(length uint16, levels map[uint]struct{}) (string, float64, error) {
	var characters []string
	b := make([]rune, length)

	if length < 1 {
		return "", 0, errors.New("password length must be equal to or higher than 1")
	}

	// Append characters to a slice
	for key := range levels {
		if key > 8 {
			return "", 0, errors.New("password level must be equal to or lower than 8")
		}

		switch key {
		case 1:
			characters = append(characters, lowerCase)
		case 2:
			characters = append(characters, upperCase)
		case 3:
			characters = append(characters, digits)
		case 4:
			characters = append(characters, space)
		case 5:
			characters = append(characters, brackets)
		case 6:
			characters = append(characters, points)
		case 7:
			characters = append(characters, special)
		case 8:
			characters = append(characters, extended)
		}
	}

	// Join slice parts and convert to rune
	join := strings.Join(characters, "")
	pool := []rune(join)

	entropy := calculateEntropy(length, len(pool))

	for i := range b {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
		b[i] = pool[randInt.Int64()]
	}

	password := string(b)

	return password, entropy, nil
}

// Entropy is a measure of what the password could have been so it does not really relate to the password
// itself, but to the selection process.
func calculateEntropy(length uint16, pool int) float64 {
	pow := math.Pow(float64(pool), float64(length))
	entropy := math.Log2(pow)
	// nAttempts := pow / 2 -> this is a generic way to measure the number of attempts, human choices are left out

	return entropy
}
