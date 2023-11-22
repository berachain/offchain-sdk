package types

import "encoding/asn1"

// asn1EcPublicKey is a struct to parse ECDSA public key from KMS.
type Asn1EcPublicKey struct {
	EcPublicKeyInfo Asn1EcPublicKeyInfo
	PublicKey       asn1.BitString
}

// asn1EcPublicKeyInfo is a struct to parse ECDSA public key from KMS.
type Asn1EcPublicKeyInfo struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

// asn1EcSig is a struct to parse ECDSA signature from KMS.
type Asn1EcSig struct {
	R asn1.RawValue
	S asn1.RawValue
}
