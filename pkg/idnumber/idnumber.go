// Package idnumber provides some primitives for parsing/understanding and even creating South African ID Numbers
package idnumber

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/theplant/luhn"
)

const (
	minFemale = 0
	maxFemale = 4999
	maxMale   = 9999
	minMale   = 5000
)

func randomFemale() GenderCode {
	return GenderCode(minFemale + rand.Intn(4999))
}

func randomMale() GenderCode {
	return GenderCode(minMale + rand.Intn(4999))
}

// Citizenship is an enumeration of either South African Citizens or Permanent Residents
type Citizenship int

const (
	// Citizen is the default South African Citizen
	Citizen Citizenship = iota
	// PermanentResident is the alternative
	PermanentResident
)

func (c Citizenship) String() string {
	switch c {
	case Citizen:
		return "citizen"
	case PermanentResident:
		return "permanent resident"
	default:
		return "undefined citizenship"
	}
}

// GenderCode saves the number ranges that represents "sex" (Used Gender here, perhaps incorrectly)
type GenderCode int

// IDNumber is the structure that contains the meaning behind the long South African ID number string
type IDNumber struct {
	birthdate   time.Time
	gender      GenderCode
	citizenship Citizenship
	luhnValue   int
}

var (
	ErrIncorrectIDStringLength = fmt.Errorf("incorrect ID string length")
	ErrIncorrectGenderRange    = fmt.Errorf("incorrect gender range")
	ErrInvalidLuhnNumber       = fmt.Errorf("invalid luhn number")
)

func (g GenderCode) String() string {
	if g >= minFemale && g <= maxFemale {
		return "female"
	} else if g >= minMale && g <= maxMale {
		return "male"
	}
	return "undefined"
}

// Citizenship provides the person's citizen status
func (id IDNumber) Citizenship() Citizenship {
	return id.citizenship
}

func (id IDNumber) String() string {
	return fmt.Sprintf("%s%0.4d%d8%d", id.birthdate.Format("060102"), id.gender, id.citizenship, id.luhnValue)
}

// Explain prints out a more verbose explanation of what the ID number means
func (id IDNumber) Explain() string {
	return fmt.Sprintf("Birthdate: %s %s %s luhn checksum = %d", id.birthdate.Format("2 January '06"), id.gender, id.citizenship, id.luhnValue)
}

// ConfigOption provides a way to configure a new IDNumber object
type ConfigOption func(id *IDNumber) error

// SetDate provides a way of setting the date of a person's birth
func SetDate(day int, month time.Month, year int) ConfigOption {
	return func(idNumber *IDNumber) error {
		idNumber.birthdate = time.Date(year, month, day, 0, 0, 0, 0, &time.Location{})
		return nil
	}
}

// SetGender sets the sex of the person
func SetGender(gender GenderCode) ConfigOption {
	return func(idNumber *IDNumber) error {
		idNumber.gender = gender
		return nil
	}
}

// SetRandomMale creates a random number that will indicate male IDNumbers
func SetRandomMale() ConfigOption {
	code := randomMale()
	return SetGender(code)
}

// SetRandomFemale creates a random number that will indicate female IDNumbers
func SetRandomFemale() ConfigOption {
	code := randomFemale()
	return SetGender(code)
}

func setCitizenship(citizenship Citizenship) ConfigOption {
	return func(idNumber *IDNumber) error {
		idNumber.citizenship = citizenship
		return nil
	}
}

// SetCitizen sets the IDNumber to an natural South African Citizen
func SetCitizen() ConfigOption {
	return setCitizenship(Citizen)
}

// SetResident sets the IDNumber to a Permanent Resident
func SetResident() ConfigOption {
	return setCitizenship(PermanentResident)
}

// SetFromString takes an id number string and creates an IDNumber from it
func SetFromString(id string) ConfigOption {
	// YYMMDD GGGG C  8 L
	//                |
	//             legacy bit, always there, ignore.
	const (
		dateIndex     = 0
		dateLength    = 6
		genderIndex   = 6
		genderLength  = 4
		citizenIndex  = 10
		citizenLength = 1
		luhnIndex     = 12
		luhnLength    = 1
		idLength      = 13
	)

	return func(idNumber *IDNumber) error {
		if len(id) != idLength {
			return ErrIncorrectIDStringLength
		}
		dateString := id[dateIndex : dateIndex+dateLength]
		t, err := time.Parse("060102", dateString)
		if err != nil {
			return err
		}
		idNumber.birthdate = t

		genderCode, err := strconv.ParseInt(id[genderIndex:genderIndex+genderLength], 10, 32)
		if err != nil {
			return err
		}
		idNumber.gender = GenderCode(genderCode)

		citizenship, err := strconv.ParseInt(id[citizenIndex:citizenIndex+citizenLength], 10, 32)
		if err != nil {
			return err
		}
		idNumber.citizenship = Citizenship(citizenship)

		luhnNumber, err := strconv.ParseInt(id[luhnIndex:luhnIndex+luhnLength], 10, 32)
		if err != nil {
			return err
		}
		idNumber.luhnValue = int(luhnNumber)
		return nil
	}
}

// NewIDNumber builds a new IDNUmber
func NewIDNumber(configOptions ...ConfigOption) (*IDNumber, error) {
	idNumber := &IDNumber{
		luhnValue: -1,
	}
	for _, configOption := range configOptions {
		err := configOption(idNumber)
		if err != nil {
			return nil, err
		}
	}
	partialString := fmt.Sprintf("%s%0.4d%d8", idNumber.birthdate.Format("060102"), idNumber.gender, idNumber.citizenship)
	partialID, err := strconv.ParseInt(partialString, 10, 64)
	if err != nil {
		return nil, err
	}

	luhnValue := luhn.CalculateLuhn(int(partialID))
	if idNumber.luhnValue > 0 {
		if idNumber.luhnValue != luhnValue {
			return nil, ErrInvalidLuhnNumber
		}
	} else {
		idNumber.luhnValue = luhnValue
	}
	return idNumber, nil
}

// NewID builds a new ID number with a simple builder
func NewID(day int, month time.Month, year int, gender GenderCode, citizenship Citizenship) (*IDNumber, error) {
	return NewIDNumber(
		SetDate(day, month, year),
		SetGender(gender),
		setCitizenship(citizenship),
	)
}

// RandomIDNumber *should* create a random valid IDNumber
func RandomIDNumber() (*IDNumber, error) {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	randomDate := time.Unix(sec, 0)

	var genderOption ConfigOption
	if rand.Intn(100) > 50 {
		genderOption = SetRandomFemale()
	} else {
		genderOption = SetRandomMale()
	}

	var citizenOption ConfigOption
	if rand.Intn(100) > 50 {
		citizenOption = SetCitizen()
	} else {
		citizenOption = SetResident()
	}
	retID, err := NewIDNumber(
		SetDate(randomDate.Day(), randomDate.Month(), randomDate.Year()),
		genderOption,
		citizenOption,
	)
	if err != nil {
		return nil, err
	}
	return retID, nil
}
