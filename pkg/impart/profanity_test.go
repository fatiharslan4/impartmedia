package impart_test

import (
	"testing"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/stretchr/testify/require"
)

func TestProfanityTest(t *testing.T) {
	impart.TestInitProfanityDetector()

	word := "bad ass"
	resultString, err := impart.ProfanityDetector.CensorWord(word)
	if err != nil {
		panic(err)
	}

	require.Equal(t, resultString, "bad ***")
}

func TestBadWordFirstLetterKept(t *testing.T) {
	word := "bitch"
	resultString, err := impart.ProfanityDetector.CensorWord(word)
	if err != nil {
		panic(err)
	}

	require.Equal(t, resultString, "*****")
}

func TestBadLongWord(t *testing.T) {
	word := "this is a long bitch string a55hole have bad fucking words"
	resultString, err := impart.ProfanityDetector.CensorWord(word)
	if err != nil {
		panic(err)
	}

	require.Equal(t, resultString, "this is a long ***** string ******* have bad ******* words")
}
