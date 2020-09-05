package entry

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
	"time"
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

// Entry represents a record.
type Entry struct {
	Title    string    `json:",omitempty"`
	Username string    `json:",omitempty"`
	Password string    `json:",omitempty"`
	URL      string    `json:",omitempty"`
	Expires  time.Time `json:",omitempty"`
}

// New creates a new entry.
func New(title, username, password, url string, expires time.Time) *Entry {
	return &Entry{
		Title:    title,
		Username: username,
		Password: password,
		URL:      url,
		Expires:  expires,
	}
}

// GeneratePassword generates a random password with the length and format given.
func GeneratePassword(length uint16, levels map[uint]struct{}) (string, float64, error) {
	var characters []string
	b := make([]rune, length)

	if length < 1 {
		return "", 0, errors.New("Password length must be equal to or higher than 1")
	}

	// Append characters to a slice
	for key := range levels {
		if key > 8 {
			return "", 0, errors.New("Password level must be equal to or lower than 8")
		}
		if key == 1 {
			characters = append(characters, lowerCase)
		}
		if key == 2 {
			characters = append(characters, upperCase)
		}
		if key == 3 {
			characters = append(characters, digits)
		}
		if key == 4 {
			characters = append(characters, space)
		}
		if key == 5 {
			characters = append(characters, brackets)
		}
		if key == 6 {
			characters = append(characters, points)
		}
		if key == 7 {
			characters = append(characters, special)
		}
		if key == 8 {
			characters = append(characters, extended)
		}
	}

	// Join slice parts and convert to rune
	join := strings.Join(characters, "")
	chunk := []rune(join)

	entropy := calculateEntropy(length, len(join))

	for i := range b {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chunk))))
		b[i] = chunk[randInt.Int64()]
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
