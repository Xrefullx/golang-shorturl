package handlers

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestGenerateLink(t *testing.T) {
	type link struct {
		searchLink string
		encod      string
	}
	TestArray := []struct {
		name string
		link link
	}{
		{
			name: "test 1",
			link: link{
				searchLink: "https://yandex.ru/search/?text=autotest+go&lr=62&clid=1882628",
				encod:      "5",
			},
		},
		{
			name: "test 2",
			link: link{
				searchLink: "https://ya.ru",
				encod:      "9",
			},
		},
	}
	for _, x := range TestArray {
		t.Run(x.name, func(t *testing.T) {
			g := GenerateLink(x.link.searchLink, x.link.encod)
			assert.False(t, len(g) == 0 || len(g) > 10)
			regex := "^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$"
			statusOk, _ := regexp.MatchString(regex, g)
			assert.True(t, statusOk)
		})
	}
}
