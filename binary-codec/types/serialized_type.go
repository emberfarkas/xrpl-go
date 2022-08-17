package types

type SerializedType interface {
	SerializeJson(json any) ([]byte, error)
}

func GetSerializedType(t string) SerializedType {
	switch t {
	case "UInt8":
		return &UInt8{}
	case "UInt16":
		return &UInt16{}
	case "UInt32":
		return &UInt32{}
	case "UInt64":
		return &UInt64{}
	case "Hash256":
		return &Hash256{}
	case "AccountID":
		return &AccountID{}
	}
	return nil
}
