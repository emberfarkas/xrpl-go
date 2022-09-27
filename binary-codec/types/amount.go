package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Amount struct{}

const (
	MinIOUExponent  = -96
	MaxIOUExponent  = 80
	MaxIOUPrecision = 16
	MinIOUMantissa  = 1e16
	MaxIOUMantissa  = 1e17 - int(1)

	NotXRPBitMask            = 0x80
	PosSignBitMask           = 0x4000000000000000
	ZeroCurrencyAmountHex    = 0x8000000000000000
	NativeAmountByteLength   = 8
	CurrencyAmountByteLength = 48

	MinXRP   = 1e-6
	MaxDrops = 1e17 // 100 billion XRP in drops aka 10^17

	AllowedIOUCharacters = "0123456789.-eE" // going to do this with Regex
)

type BigDecimal struct {
	Scale         int
	Precision     int
	UnscaledValue string
	Sign          string
}

func (a *Amount) SerializeJson(value any) ([]byte, error) {

	return nil, nil
}

// Creates a new custom BigDecimal object from a value string
func NewBigDecimal(value string) (*BigDecimal, error) {

	bigDecimal := new(BigDecimal)
	var err error

	if ContainsInvalidCharacters(value) {
		return nil, errors.New("value contains invalid characters: only '0-9' '.' '-' 'e' and 'E' are allowed")
	}
	if strings.HasPrefix(value, "-") { // if the value is negative, set the sign to negative
		bigDecimal.Sign = "-"
	}

	value = strings.ToLower(strings.TrimPrefix(value, "-")) // remove the sign from the value

	prefix, suffix, ePresent := strings.Cut(value, "e") // split the value into prefix and suffix at the 'e' character, if present

	if strings.Contains(prefix, "-") { // if the value still has a minus sign in the prefix, return an error
		return nil, errors.New("value contains multiple '-' characters, excluding the exponent sign")
	}
	if strings.Contains(prefix, "e") || strings.Contains(suffix, "e") { // if the prefix or suffix still contains an 'e' character, return an error
		return nil, errors.New("value contains multiple 'e' or 'E' characters")
	}

	emptyCheck := strings.Trim(prefix, "0") // remove all zeros from the prefix

	if emptyCheck == "" || emptyCheck == "." { // if the prefix is empty or just a decimal point, set everything to 0 or "" and return
		bigDecimal.Scale = 0
		bigDecimal.Precision = 0
		bigDecimal.UnscaledValue = ""
		bigDecimal.Sign = ""
		return bigDecimal, nil
	}

	containsDecimal, decimalErr := containsDecimal(value) // check if the value contains a SINGLE decimal point

	if decimalErr != nil {
		return nil, decimalErr
	}

	decimalPrefix, decimalSuffix, _ := strings.Cut(prefix, ".") // split the prefix into decimal prefix and decimal suffix at the '.' character, if present

	if ePresent { // if the value contains an 'e' character
		bigDecimal.Scale, err = strconv.Atoi(suffix) // convert the suffix to an int, which is the scale
		if err != nil {
			return nil, err
		}

		if containsDecimal { // if the value contains a SINGLE decimal point
			decimalSuffixNoTrailingZeros := strings.TrimRight(decimalSuffix, "0")         // remove trailing zeros from the decimal suffix
			decimalPrefixNoLeadingZeros := strings.TrimLeft(decimalPrefix, "0")           // remove leading zeros from the decimal prefix
			bigDecimal.Scale = bigDecimal.Scale - len(decimalSuffixNoTrailingZeros)       // subtract the length of the decimal suffix from the scale
			bigDecimal.UnscaledValue = strings.Trim((decimalPrefix + decimalSuffix), "0") // remove leading and trailing zeros from the concatenated decimal prefix and decimal suffix to get the unscaled value

			if decimalSuffixNoTrailingZeros == "" { // if the decimal suffix is empty because it only contained trailing zeros
				bigDecimal.Scale = bigDecimal.Scale + (len(decimalPrefixNoLeadingZeros) - len(bigDecimal.UnscaledValue)) // add the difference between the length of the decimal prefix and the length of the unscaled value, to the scale
			}

		} else if !containsDecimal { // if the value does not contain a SINGLE decimal point
			prefixNoTrailingZeros := strings.Trim(prefix, "0")                                             // remove trailing zeros from the prefix
			prefixNoLeadingZeros := strings.TrimLeft(prefix, "0")                                          // remove leading zeros from the prefix
			bigDecimal.Scale = bigDecimal.Scale + (len(prefixNoLeadingZeros) - len(prefixNoTrailingZeros)) // add the difference between the length of the prefix and the length of the prefix with trailing zeros removed, to the scale
			bigDecimal.UnscaledValue = prefixNoTrailingZeros                                               // set the unscaled value to the prefix with trailing zeros removed
		} else {
			return nil, err
		}
	}

	if !ePresent { // if the value does not contain an 'e' character

		if containsDecimal && decimalErr == nil { // if the value contains a SINGLE decimal point
			decimalSuffixNoTrailingZeros := strings.TrimRight(decimalSuffix, "0")                        // remove trailing zeros from the decimal suffix
			decimalPrefixNoLeadingZeros := strings.TrimLeft(decimalPrefix, "0")                          // remove leading zeros from the decimal prefix
			bigDecimal.Scale = -len(decimalSuffixNoTrailingZeros)                                        // set the scale to the negative of the length of the decimal suffix with trailing zeros removed
			bigDecimal.UnscaledValue = strings.Trim((decimalPrefix + decimalSuffixNoTrailingZeros), "0") // remove leading and trailing zeros from the concatenated decimal prefix and decimal suffix with trailing zeros removed to get the unscaled value

			if decimalSuffixNoTrailingZeros == "" { // if the decimal suffix is empty because it only contained trailing zeros
				bigDecimal.Scale = len(decimalPrefixNoLeadingZeros) - len(bigDecimal.UnscaledValue) // set the scale to the difference between the length of the decimal prefix and the length of the unscaled value
			}

		} else if !containsDecimal && decimalErr == nil { // if the value does not contain a SINGLE decimal point
			decimalPrefixNoTrailingZeros := strings.TrimRight(prefix, "0")             // remove trailing zeros from the prefix
			bigDecimal.Scale = len(prefix) - len(decimalPrefixNoTrailingZeros)         // set the scale to the difference between the length of the prefix and the length of the prefix with trailing zeros removed
			bigDecimal.UnscaledValue = strings.Trim(decimalPrefixNoTrailingZeros, "0") // remove leading and trailing zeros from the prefix with trailing zeros removed to get the unscaled value
		} else {
			return nil, err
		}
	}

	bigDecimal.Precision = len(bigDecimal.UnscaledValue) // set the precision to the length of the unscaled value

	return bigDecimal, nil
}

// validates the format of an XRP amount value
// XRP values shouldn't contain a decimal point BECAUSE they are represented as integers as drops
func VerifyXrpValue(value string) error {

	containsDecimal, err := containsDecimal(value)

	if err != nil {
		return err
	} else if containsDecimal {
		return errors.New("XRP value must not contain a decimal")
	}

	decimal := new(big.Float)
	decimal, ok := decimal.SetString(value) // bigFloat for precision

	if !ok {
		return errors.New("failed to convert string to big.Float")
	}

	if decimal.Sign() == 0 {
		return nil
	}

	if decimal.Cmp(big.NewFloat(MinXRP)) == -1 || decimal.Cmp(big.NewFloat(MaxDrops)) == 1 {
		return errors.New("XRP value is an invalid XRP amount")
	}

	return nil
}

func (b *BigDecimal) GetExponent() int {
	return b.Precision - b.Scale - 1
}

// validates the format of an issued currency amount value
func VerifyIOUValue(value string) error {

	bigDecimal, err := NewBigDecimal(value)
	exp := bigDecimal.GetExponent()

	if err != nil {
		return err
	}

	decimalValue, ok := new(big.Float).SetString(value) // bigFloat for precision

	if !ok {
		return errors.New("failed to convert string to big.Float")
	}

	if decimalValue.Sign() == 0 {
		return nil
	}

	if bigDecimal.Precision > MaxIOUPrecision {
		return errors.New("IOU value is an invalid IOU amount - precision is too large > 16") // if the precision is greater than 16, return an error
	}
	if exp < MinIOUExponent {
		return errors.New("IOU value is an invalid IOU amount - exponent is out of range") // if the scale is less than -96 or greater than 80, return an error
	}
	if exp > MaxIOUExponent {
		return errors.New("IOU value is an invalid IOU amount - exponent is out of range") // if the scale is less than -96 or greater than 80, return an error
	}

	return err
}

// Serializes an XRP amount value
func SerializeXrpAmount(value string) ([]byte, error) {

	if VerifyXrpValue(value) != nil {
		return nil, VerifyXrpValue(value)
	}

	val, err := strconv.ParseUint(value, 10, 64)

	if err != nil {
		return nil, err
	}

	valWithPosBit := val | PosSignBitMask
	valBytes := make([]byte, NativeAmountByteLength)

	binary.BigEndian.PutUint64(valBytes, uint64(valWithPosBit))

	return valBytes, nil
}

// Returns true if this amount is a "native" XRP amount - first bit in first byte set to 0 for native XRP
func isNative(value []byte) bool {
	fmt.Printf("%08b", value)
	x := []byte(value)[0]&NotXRPBitMask == 0 // & bitwise operator returns 1 if both first bits are 1, otherwise 0
	return x
}

// Determines if this AmountType is positive - 2nd bit in 1st byte is set to 1 for positive amounts
func isPositive(value []byte) bool {
	fmt.Printf("%08b", value)
	x := []byte(value)[0]&0x40 > 0
	return x
}

// XRP values shouldn't contain a decimal point BECAUSE they are represented as integers as drops
// returns true if the string contains a SINGLE decimal point character
func containsDecimal(s string) (bool, error) {
	decCount := strings.Count(s, ".") // count the number of decimal points
	if decCount > 1 {
		return true, errors.New("invalid - string contains more than one decimal point")
	}
	return strings.Contains(s, "."), nil
}

// Checks if a value string contains invalid characters - returns true if it does
func ContainsInvalidCharacters(value string) bool {
	for _, char := range value {
		if !strings.Contains(AllowedIOUCharacters, strings.ToLower(string(char))) { // if the character is not in the allowed characters list return true
			return true
		}
	}
	return false
}