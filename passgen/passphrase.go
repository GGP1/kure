package passgen

import (
	"math"
	"strings"
)

var (
	vowels    = []string{"a", "e", "i", "o", "u"}
	constants = []string{"b", "c", "f", "g", "h", "j", "k", "l", "m", "n",
		"p", "q", "r", "s", "t", "v", "w", "x", "y", "z"}
)

// Passphrase represents a sequence of words required for access to th system.
type Passphrase struct {
	Phrase    string
	Length    uint64
	Separator string
}

// Generate generates a random passphrase with the length given.
//
// Words in the passphrase are separated by the separator character specified.
//
// We don't use a word list because it would make the job easier for the potential attacker.
func (p *Passphrase) Generate() (string, error) {
	passphrase := make([]string, 0, p.Length)

	// Running goroutines with a waitgroup on each word/syllable doesn't have a significant improvement
	for i := 0; i < int(p.Length); i++ {
		// 3 (min), 12 (max) length of a word in this algorithm
		wordLength := randInt(10) + 3

		syllables := make([]string, 0, wordLength)

		for j := 0; j < wordLength; j++ {
			// Take a number from 0 to 10
			// 0 to 3 add a vowel, 4 to 10 add a constant
			if randInt(11) <= 3 {
				syllables = append(syllables, vowels[randInt(len(vowels))])
			} else {
				syllables = append(syllables, constants[randInt(len(constants))])
			}
		}

		// Join syllables and append the word to the slice
		word := strings.Join(syllables, "")
		passphrase = append(passphrase, word)
	}

	// Join all words with the separator specified
	p.Phrase = strings.Join(passphrase, p.Separator)

	return p.Phrase, nil
}

// Entropy is a measure of what the passphrase could have been so it does not really relate to the passphrase
// itself, but to the selection process. It takes the length of the passphrase and all the possible characters
// from the pool and returns the entropy bits of the passphrase.
func (p *Passphrase) Entropy() float64 {
	poolLength := len(vowels) + len(constants) + len(p.Separator)
	phraseLength := len(p.Phrase)

	pow := math.Pow(float64(poolLength), float64(phraseLength))
	entropy := math.Log2(pow)

	// This is a generic way to measure the number of attempts, dictionary attacks, social engineering
	// and other techniques are left out of consideration
	// nAttempts := pow / 2

	return entropy
}
