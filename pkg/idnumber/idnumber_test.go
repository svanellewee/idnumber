package idnumber_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/svanellewee/idnumber/pkg/idnumber"
)

func TestSouthAfricanIDNumber(t *testing.T) {
	// Checking if a string is a valid luhn
	dude, err := idnumber.NewID(9, 7, 1981, 5005, idnumber.Citizen)
	assert.Nil(t, err)
	assert.Equal(t, "8107095005083", dude.String())
	fmt.Printf("%v..\n", dude)
	for i := 0; i < 10; i++ {
		id, err := idnumber.RandomIDNumber()
		assert.Nil(t, err)
		fmt.Printf("%s %s\n", id, id.Explain())
	}
	idValue, err := idnumber.NewIDNumber(idnumber.SetFromString("8107095005083"))
	assert.Nil(t, err)
	assert.Equal(t, dude.String(), idValue.String())
}
