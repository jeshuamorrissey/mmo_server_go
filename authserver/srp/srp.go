package srp

import (
	"crypto/sha1"
	"math/big"
)

const (
	// G is the SRP Generator; the base of many mathematical expressions.
	G uint8 = 7

	// K is the SRP Verifier Scale Factor; used to scale the verifier which
	// is stored in the database.
	K uint8 = 3
)

// N is the SRP Modulus; all operations are performed in base N.
func N() *big.Int {
	n := big.NewInt(0)
	n.SetString("62100066509156017342069496140902949863249758336000796928566441170293728648119", 10)
	return n
}

// GenerateSalt generates a random salt.
func GenerateSalt() *big.Int {
	// TODO(jeshua): Make this a random number.
	s := big.NewInt(0)
	s.SetString("66759882342950727220130969932663635787137805713109467932708165413389947953699", 10)
	return s
}

func _H(parts ...[]byte) []byte {
	hash := sha1.New()
	for _, part := range parts {
		hash.Write(reverse(part))
	}

	return reverse(hash.Sum(nil))
}

// GenerateVerifier will generate a hash of the account name, password and salt
// which can be used as the SRP verifier.
func GenerateVerifier(accountName, password string, salt *big.Int) *big.Int {
	x := big.NewInt(0)
	x.SetBytes(_H(salt.Bytes(), _H(reverse([]byte(accountName)), []byte(":"), reverse([]byte(password)))))

	g := big.NewInt(int64(G))
	return g.Exp(g, x, N())
}

// GenerateEphemeral generates a (private, public) ephemeral pair given a user's
// verifier.
func GenerateEphemeral(v *big.Int) (*big.Int, *big.Int) {
	// TODO(jeshua): Make this a random number.
	b := big.NewInt(0)
	b.SetString("3679141816495610969398422835318306156547245306", 10)

	g := big.NewInt(int64(G))

	B := big.NewInt(0)
	B.Mul(v, big.NewInt(3))
	B.Add(B, g.Exp(g, b, N()))
	B.Mod(B, N())

	return b, B
}

func padBigIntBytes(data []byte, nBytes int) []byte {
	if len(data) > nBytes {
		return data[:nBytes]
	}

	currSize := len(data)
	for i := 0; i < nBytes-currSize; i++ {
		data = append([]byte{'\x00'}, data...)
	}

	return data
}

func interleave(S *big.Int) *big.Int {
	T := padBigIntBytes(reverse(S.Bytes()), 32)

	G := make([]byte, 16)
	H := make([]byte, 16)
	for i := 0; i < 16; i++ {
		G[i] = T[i*2]
		H[i] = T[i*2+1]
	}

	G = reverse(_H(reverse(G)))
	H = reverse(_H(reverse(H)))

	K := make([]byte, 0)
	for i := 0; i < 20; i++ {
		K = append(K, G[i], H[i])
	}

	KInt := big.NewInt(0)
	KInt.SetBytes(reverse(K))
	return KInt
}

func reverse(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return data
}

// CalculateSessionKey takes as input the client's proof and calculates the
// persistent session key.
func CalculateSessionKey(A, B, b, v, s *big.Int, accountName string) (*big.Int, *big.Int) {
	u := big.NewInt(0)
	u.SetBytes(_H(A.Bytes(), B.Bytes()))

	S := big.NewInt(0)
	S.Exp(v, u, N())
	S.Mul(S, A)
	S.Exp(S, b, N())

	K := interleave(S)

	NHash := big.NewInt(0)
	NHash.SetBytes(_H(N().Bytes()))

	gHash := big.NewInt(0)
	gHash.SetBytes(_H(big.NewInt(int64(G)).Bytes()))
	gHash.Xor(gHash, NHash)

	M := big.NewInt(0)
	M.SetBytes(_H(gHash.Bytes(), _H(reverse([]byte(accountName))), s.Bytes(), A.Bytes(), B.Bytes(), K.Bytes()))
	return K, M
}

// CalculateServerProof will calculate a proof to send back to the client so they
// know we are a legit server.
func CalculateServerProof(A, M, K *big.Int) *big.Int {
	proof := big.NewInt(0)
	proof.SetBytes(_H(A.Bytes(), M.Bytes(), K.Bytes()))
	return proof
}
