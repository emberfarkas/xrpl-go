package types

import "fmt"

const HashLengthBytes = 32

type ErrInvalidVector256Type struct {
	Got string
}

func (e *ErrInvalidVector256Type) Error() string {
	return fmt.Sprintf("Invalid type to construct Vector256 from. Expected []string, got %v", e.Got)
}

type Vector256 struct{}

func (v *Vector256) SerializeJson(json any) ([]byte, error) {

	if _, ok := json.([]string); !ok {
		return nil, &ErrInvalidVector256Type{fmt.Sprintf("%T", json)}
	}

	b, err := vector256FromValue(json.([]string))

	if err != nil {
		return nil, err
	}

	return b, nil

}

func vector256FromValue(value []string) ([]byte, error) {
	b := make([]byte, 0)
	for _, s := range value {
		hash256, err := NewHash256().SerializeJson(s)

		if err != nil {
			return nil, err
		}

		b = append(b, hash256...)

	}
	return b, nil
}