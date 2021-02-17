package models

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHiveDistributions_Sort(t *testing.T) {
	length := 10
	hds := make(HiveDistributions, length, length)
	for i, _ := range hds {
		hds[i].DisplayValue = strconv.Itoa(i)
		hds[i].SortValue = length - i
	}

	assert.Len(t, hds, length)
	assert.Equal(t, 1, hds[length-1].SortValue)
	assert.Equal(t, "0", hds[0].DisplayValue)

	assert.False(t, hds.IsSorted())
	hds.Sort()

	assert.Len(t, hds, length)
	assert.Equal(t, 10, hds[length-1].SortValue)
	assert.Equal(t, "9", hds[0].DisplayValue)
	assert.True(t, hds.IsSorted())
}

func TestHiveDistributions_Pop(t *testing.T) {
	length := 10
	hds := make(HiveDistributions, length, length)
	for i, _ := range hds {
		hds[i].DisplayValue = strconv.Itoa(i)
		hds[i].SortValue = length - i
	}

	assert.Len(t, hds, length)
	assert.Equal(t, 1, hds[length-1].SortValue)
	assert.Equal(t, "0", hds[0].DisplayValue)
	assert.Equal(t, "5", hds[5].DisplayValue)

	hds.Pop(5)

	assert.Len(t, hds, 9)
	assert.Equal(t, "6", hds[5].DisplayValue)
	assert.Equal(t, "4", hds[4].DisplayValue)
	assert.Equal(t, "9", hds[len(hds)-1].DisplayValue)

	hds.Pop(8)
	assert.Len(t, hds, 8)
	assert.Equal(t, "6", hds[5].DisplayValue)
	assert.Equal(t, "4", hds[4].DisplayValue)
	assert.Equal(t, "8", hds[len(hds)-1].DisplayValue)
	assert.Equal(t, "0", hds[0].DisplayValue)

	hds.Pop(0)
	assert.Len(t, hds, 7)
	assert.Equal(t, "6", hds[4].DisplayValue)
	assert.Equal(t, "4", hds[3].DisplayValue)
	assert.Equal(t, "8", hds[len(hds)-1].DisplayValue)
	assert.Equal(t, "1", hds[0].DisplayValue)
}
