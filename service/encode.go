package service

import (
	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

func encodingFromRPC(enc Encoding) (keys.Encoding, error) {
	switch enc {
	case Base62:
		return keys.Base62, nil
	case Base58:
		return keys.Base58, nil
	case Base32:
		return keys.Base32, nil
	case Hex:
		return keys.Hex, nil
	case Base64:
		return keys.Base64, nil
	case Saltpack:
		return keys.Saltpack, nil
	case BIP39:
		return keys.BIP39, nil
	default:
		return keys.Base62, errors.Errorf("invalid encoding")
	}
}

func encodingToRPC(enc string) (Encoding, error) {
	switch enc {
	case "base62":
		return Base62, nil
	case "base58":
		return Base58, nil
	case "base32":
		return Base32, nil
	case "hex":
		return Hex, nil
	case "base64":
		return Base64, nil
	case "saltpack":
		return Saltpack, nil
	case "bip39":
		return BIP39, nil
	default:
		return Base62, errors.Errorf("invalid encoding")
	}
}
