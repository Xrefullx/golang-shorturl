package memory

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestMaps_Get(t *testing.T) {
	type body struct {
		short string
	}
	testArray := []struct {
		testNumber   string
		mp           Maps
		testbody     body
		shouldReturn string
		err          bool
	}{
		{
			testNumber: "1",
			mp: Maps{urlMap: map[string]string{
				"test": "https://ya.ru",
			},
				mutex: &sync.RWMutex{},
			},
			testbody: body{short: ""},
			err:      true,
		},
		{
			testNumber: "2",
			mp: Maps{urlMap: map[string]string{
				"test": "https://ya.ru",
			},
				mutex: &sync.RWMutex{},
			},
			testbody:     body{short: "test123"},
			shouldReturn: "",
			err:          false,
		},
		{
			testNumber: "3",
			mp: Maps{urlMap: map[string]string{
				"test": "https://ya.ru",
			},
				mutex: &sync.RWMutex{},
			},
			testbody:     body{short: "test123"},
			shouldReturn: "",
			err:          false,
		},
	}
	for _, x := range testArray {
		t.Run(x.testNumber, func(t *testing.T) {
			get, err := x.mp.Get(x.testbody.short)
			if !x.err {
				require.NoError(t, err)
				assert.Equal(t, x.shouldReturn, get)
				return
			}
			assert.Error(t, err)
		})

	}
}

func TestMaps_Save(t *testing.T) {
	type body struct {
		search string
	}
	testArray := []struct {
		testNumber string
		mp         Maps
		testbody   body
		err        bool
	}{
		{
			testNumber: "1",
			mp: Maps{urlMap: map[string]string{
				"test": "https://ya.ru",
			}, mutex: &sync.RWMutex{},
			},
			testbody: body{search: ""},
			err:      true,
		},
		{
			testNumber: "2",
			mp: Maps{urlMap: map[string]string{
				"test": "https://ya.ru",
			}, mutex: &sync.RWMutex{},
			},
			testbody: body{search: "https://yandex.ru/pogoda"},
			err:      false,
		},
	}
	for _, x := range testArray {
		t.Run(x.testNumber, func(t *testing.T) {
			get, err := x.mp.Get(x.testbody.search)
			if !x.err {
				require.NoError(t, err)
				assert.Equal(t, x.mp.urlMap[get], get)
				return
			}
			assert.Error(t, err)
		})
	}
}
