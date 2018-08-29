package crypto

import (
	"crypto/sha512"
	"io"
	"strconv"
	"crypto/hmac"

	cryptorand "crypto/rand"
	"github.com/lomocoin/lws/internal/gateway/crypto/edwards25519"
	"golang.org/x/crypto/ripemd160"
)

const (
	SeedSize = 32
	PublicKeySize = 32
	PrivateKeySize = 32
	ApiKeySize = 32
)

// PublicKey is the type of Ed25519 public keys.
type PublicKey [32]byte

// PrivateKey is the type of Ed25519 private keys. It implements crypto.Signer.
type PrivateKey [32]byte

// ApiKey is an exchange key.
type ApiKey [32]byte

func GenerateKeyPair(rand io.Reader) (PublicKey, PrivateKey){
	if rand == nil {
		rand = cryptorand.Reader
	}

	seed := make([]byte, SeedSize)
	if _, err := io.ReadFull(rand, seed); err != nil {
		// return nil, err
		panic("ed25519: seed")
	}

	if l := len(seed); l != SeedSize {
		panic("ed25519: bad seed length: " + strconv.Itoa(l))
	}

	digest := sha512.Sum512(seed)
	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	var A edwards25519.ExtendedGroupElement
	var hBytes [32]byte

	copy(hBytes[:], digest[:])
	edwards25519.GeScalarMultBase(&A, &hBytes)
	var publicKeyBytes [32]byte
	A.ToBytes(&publicKeyBytes)
	var privKeyBytes [32]byte
	copy(privKeyBytes[:], digest[:32])

	return publicKeyBytes, privKeyBytes
}

func GenerateKeyApiKey(privKey *PrivateKey, pubKey *PublicKey) ApiKey {
	var apiKeyGroup edwards25519.ProjectiveGroupElement
	var pubKeyGroup edwards25519.ExtendedGroupElement
	var base [32]byte
	var apiKey [32]byte
	var pubKeyCopy [32]byte
	var privKeyCopy [32]byte
	copy(pubKeyCopy[:], pubKey[:])
	copy(privKeyCopy[:], privKey[:])
	pubKeyGroup.FromBytes(&pubKeyCopy)
	edwards25519.GeDoubleScalarMultVartime(&apiKeyGroup, &privKeyCopy, &pubKeyGroup, &base)
	apiKeyGroup.ToBytes(&apiKey)
	return apiKey
}

// sign data with apiKey by HMAC-RIPEMD-160
func SignWithApiKey(apikey []byte, message []byte)  []byte {
	h := hmac.New(ripemd160.New, apikey)
	h.Write(message)
	hbyte := h.Sum(nil)
	return hbyte
}

