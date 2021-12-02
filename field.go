package field

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

//An N-Field with Montgomery multiplication
type Field struct {
	N     *big.Int //the prime/order
	R     *big.Int
	N1    *big.Int
	R1    *big.Int
	L     uint     // len of N in bits
	rmask *big.Int // Just precomputed R-1
}

//Initialize a new N-field. N is not checked for being a prime, only for its GCD with its 2^L envelope to be 1
func NewField(prime *big.Int) (*Field, error) {
	L := len(prime.Bytes())
	mbyte := prime.Bytes()[0]
	bits := 0
	for mbyte > 0 {
		bits++
		mbyte >>= 1
	}
	L = (L-1)*8 + bits
	R := big.NewInt(2)
	R.Exp(R, big.NewInt(int64(L)), nil)
	N1 := new(big.Int)
	R1 := new(big.Int)
	z := big.NewInt(0)
	z.GCD(N1, R1, prime, R)
	if z.Cmp(big.NewInt(1)) != 0 {
		return nil, fmt.Errorf("P and R not co-prime")
	}
	N1.Neg(N1)
	if N1.Sign() < 0 || R1.Sign() < 0 {
		N1.Add(N1, R)
		R1.Add(R1, prime)
	}
	//Verify Bezout

	bezout := new(big.Int)
	bezout2 := new(big.Int)
	bezout.Sub(bezout.Mul(R, R1), bezout2.Mul(prime, N1))
	if bezout.Cmp(big.NewInt(1)) != 0 {
		return nil, fmt.Errorf("Bezou not satisfied: %v", bezout)
	}
	return &Field{prime, R, N1, R1, uint(L), new(big.Int).Sub(R, big.NewInt(1))}, nil

}

func (f *Field) REDC(T *big.Int) *big.Int {
	//TmR := new(big.Int).Mod(T, f.R)
	TmR := new(big.Int).And(T, f.rmask)
	m := new(big.Int).Mul(TmR, f.N1)
	m.Mod(m, f.R)
	t := new(big.Int).Mul(m, f.N)
	t.Add(t, T)
	//t.Div(t, f.R)
	t.Rsh(t, f.L)
	if t.Cmp(f.N) >= 0 {
		return t.Sub(t, f.N)
	}
	return t
}

func (a *Element) Mont() {
	if a.isMont {
		return
	}
	a.i.Mul(a.i, a.f.R)
	a.i.Mod(a.i, a.f.N)
	a.isMont = true
}

func (a *Element) Demont() {
	if !a.isMont {
		return
	}
	a.i.Mul(a.i, a.f.R1)
	a.i.Mod(a.i, a.f.N)
	a.isMont = false

}

type Element struct {
	i      *big.Int
	isMont bool
	f      *Field
}

func (f *Field) NewElement(i *big.Int) *Element {
	e := new(Element)
	e.i = new(big.Int).Set(i)
	e.i.Mod(i, f.N)
	e.f = f
	return e
}

func (f *Field) NewElementInt(i int) *Element {
	bi := big.NewInt(int64(i))
	return f.NewElement(bi)
}

//This is unsafe as it assumes both e1, e2 belong to the same field
//The e1 element will be brought to Montgomery form if e2 is
func (e1 *Element) Add(e2 *Element) *Element {
	if e2.isMont {
		e1.Mont()
	} else {
		e1.Demont()
	}
	e1.i.Add(e1.i, e2.i).Mod(e1.i, e1.f.N)
	return e1
}

//This is unsafe as it assumes both e1, e2 belong to the same field
//Both elements will be brought to Montgomery form
func (e1 *Element) Mul(e2 *Element) *Element {

	e1.Mont()
	e2.Mont()
	e1.i.Mul(e1.i, e2.i)
	e1.i = e1.f.REDC(e1.i)
	return e1
}

//Get the original (non-Montgomery) value
//This method does not modify the e1
func (e1 *Element) Value() *big.Int {
	h := new(big.Int)
	h.Set(e1.i)
	if e1.isMont {
		h.Mul(h, e1.f.R1)
		h.Mod(h, e1.f.N)
	}
	return h
}

func (e1 *Element) Hex() string {
	return hex.EncodeToString(e1.Value().Bytes())
}

func (e1 *Element) Clone() *Element {
	e2 := new(Element)
	e2.i = new(big.Int)
	e2.i.Set(e1.i)
	e2.isMont = e1.isMont
	e2.f = e1.f
	return e2
}

func (f *Field) RandomElement(r io.Reader) *Element {
	buf := make([]byte, len(f.R.Bytes()))
	r.Read(buf)
	i := new(big.Int)
	i.SetBytes(buf)
	return f.NewElement(i)
}
