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
	lowerCase = "abcdefghijklmnopqrstuvwxyz"                                                                                               // level 1
	upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"                                                                                               // level 2
	digits    = "0123456789"                                                                                                               // level 3
	space     = " "                                                                                                                        // level 4
	brackets  = "(){}[]<>"                                                                                                                 // level 5
	points    = ".¿?!¡,;:"                                                                                                                 // level 6
	special   = "$%&|/=*#@=~€^"                                                                                                            // level 7
	extended  = "ƒ„…†‡0ˆ‰Š‹›ŒŽ‘’“”•-+_—˜™šœžŸ¢£¤¥¦§¨©ª«¬®¯°±²³´µ¶·¸¹º»¼½¾ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõö÷øùúûüýþÿ" // level 8
)

// New creates a new entry.
func New(title, username, password, url, notes, expires string) *Entry {
	return &Entry{
		Title:    title,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
		Expires:  expires,
	}
}

// GeneratePassword generates a random password with the length and format given.
func GeneratePassword(length uint16, format []int, include string) (string, float64, error) {
	var characters []string

	password := make([]rune, length)
	levels := make(map[int]struct{}, len(format))

	for _, v := range format {
		levels[v] = struct{}{}
	}

	if length < 1 {
		return "", 0, errors.New("password length must be equal to or higher than 1")
	}

	if include != "" {
		characters = append(characters, include)
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

	// Create the pool by joining all the characeters
	pool := []rune(strings.Join(characters, ""))

	entropy := calculateEntropy(length, len(pool))

	for i := range password {
		// Generate a random number to take a character from the pool and put it into the password
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))

		password[i] = pool[randInt.Int64()]
	}

	pwd := string(password)

	return pwd, entropy, nil
}

// GeneratePassphrase generates a random passphrase with the length given.
// Words in the passphrase are separated by the separator character specified.
// We don't use a word list because it would make the job easier for the potential attacker.
func GeneratePassphrase(length int, separator string) (string, float64) {
	var (
		j         int64
		vowels    = []string{"a", "e", "i", "o", "u"}
		constants = []string{"b", "c", "f", "g", "h", "j", "k", "l", "m", "n",
			"p", "q", "r", "s", "t", "v", "w", "x", "y", "z"}
	)

	passphrase := make([]string, 0, length)

	// Running goroutines with a waitgroup on each word/syllable doesn't have a significant improvement
	for i := 0; i < length; i++ {
		// 3 (min), 12 (max) length of a word in this algorithm
		wL, _ := rand.Int(rand.Reader, big.NewInt(10))
		wordLength := wL.Add(wL, big.NewInt(3))

		syllables := make([]string, 0, wordLength.Int64())

		for j = 0; j < wordLength.Int64(); j++ {
			// randInt: take a number from 0 to 10, 4/10 is a vowel
			// randV/randC: take a random vowel/constant from the slice
			randInt, _ := rand.Int(rand.Reader, big.NewInt(11))
			randV, _ := rand.Int(rand.Reader, big.NewInt(int64(len(vowels))))
			randC, _ := rand.Int(rand.Reader, big.NewInt(int64(len(constants))))

			if randInt.Int64() <= 3 {
				syllables = append(syllables, vowels[randV.Int64()])
			} else {
				syllables = append(syllables, constants[randC.Int64()])
			}
		}

		word := strings.Join(syllables, "")
		passphrase = append(passphrase, word)
	}

	result := strings.Join(passphrase, separator)
	entropy := calculateEntropy(uint16(len(result)), len(vowels)+len(constants)+len(separator))

	return result, entropy
}

// Entropy is a measure of what the password could have been so it does not really relate to the password
// itself, but to the selection process. It takes the length of the password and all the possible characters
// from the pool and returns log2(pool^length).
func calculateEntropy(length uint16, pool int) float64 {
	pow := math.Pow(float64(pool), float64(length))
	entropy := math.Log2(pow)

	// This is a generic way to measure the number of attempts, dictionary attacks, social engineering
	// and other techniques are left out of consideration
	// nAttempts := pow / 2

	return entropy
}
