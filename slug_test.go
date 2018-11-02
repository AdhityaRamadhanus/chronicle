package chronicle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	testCases := []struct {
		Text         string
		ExpectedSlug string
	}{
		{
			"!@#!@!@!@!@ Adhitya Ramadhanus",
			"adhitya-ramadhanus",
		},
		{
			"How are you?",
			"how-are-you",
		},
		{
			"       ",
			"",
		},
		{
			" How are!@!*@&!*@&!*@!@() ",
			"how-are",
		},
		{
			"Widodo Beri Apresiasi, pada Pemain Bali United Usai Imbangi Persib",
			"widodo-beri-apresiasi-pada-pemain-bali-united-usai-imbangi-persib",
		},
	}

	for _, testCase := range testCases {
		slug := Slugify(testCase.Text)
		assert.Equal(t, slug, testCase.ExpectedSlug, "Incorrect Slug")
	}
}
