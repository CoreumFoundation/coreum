package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinPriceConversion(t *testing.T) {
	wholePart, decPart := priceToUint64s(minPrice)
	assert.Equal(t, uint64(0), wholePart)
	assert.Equal(t, uint64(1), decPart)
}

func TestMaxPriceConversion(t *testing.T) {
	wholePart, decPart := priceToUint64s(maxPrice)
	assert.Equal(t, uint64(999999999999999999), wholePart)
	assert.Equal(t, uint64(999999999999999999), decPart)
}
