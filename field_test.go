package field

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func TestNewField(t *testing.T) {

	size := 1023
	stn := big.NewInt(int64(size))
	stn.SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)
	//stn.SetString("265252859812191058636308479999999", 10)
	f, err := NewField(stn)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Field created\n", f)

	half := new(big.Int).Set(f.N)
	half.Rsh(half, 1)
	fhalf := f.NewElement(half)

	for i := -10; i < 10; i++ {

		x := f.NewElementInt(i + size/2)
		x.Add(fhalf)
		y := x.Clone()
		x.Mont()
		x.Demont()
		//fmt.Println(x.Value(), y.Value())
		if x.i.Cmp(y.i) != 0 {
			t.Errorf("Monty not reversible for %v", i)
		}
	}
	fmt.Println("Montgomery inversion passed")
	one := f.NewElementInt(1)
	fhalf.Add(fhalf)
	fhalf.Add(one)
	if fhalf.Value().Cmp(big.NewInt(0)) != 0 {
		t.Error("half is not half")
	}

	factor := one.Clone()
	next := one.Clone()

	for i := 1; i < 46; i++ {
		factor.Mul(next)
		next.Add(one)

	}
	should := big.NewInt(0)
	should.SetString("119622220865480194561963161495657715064383733760000000000", 10)
	should.Mod(should, f.N)
	if factor.Value().Cmp(should) != 0 {
		t.Error("Bad factorial:", factor.Value())
	}
	fmt.Println("Factorial ok")

	for i := 0; i < 10; i++ {
		re := f.RandomElement(rand.Reader)
		if f.N.Cmp(re.Value()) <= 0 || re.Value().Sign() <= 0 {
			t.Errorf("Bad element %v", re.Value())
		}
	}

}

func TestMon(t *testing.T) {
	stn := big.NewInt(7829)
	stn.SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)
	f, err := NewField(stn)
	if err != nil {
		t.Error(err)
	}
	for i := 3; i > 0; i-- {
		aPlain := big.NewInt(0)
		buf := make([]byte, len(f.R.Bytes()))
		rand.Read(buf)
		aPlain.SetBytes(buf)
		aPlain.Mod(aPlain, f.N)

		aMod1 := big.NewInt(1).Set(aPlain)
		aMod1.Mul(aMod1, f.R)
		aMod1.Mod(aMod1, f.N)

		aMod2 := big.NewInt(1).Set(aPlain)
		aMod2.Mul(aMod2, f.R2)

		x := f.REDC(aMod2)
		if x.Cmp(aMod1) != 0 {
			t.Error("Montgomery by REDC failed")
		}

		if f.REDC(x).Cmp(aPlain) != 0 {
			t.Error("De-Montgomery by REDC failed")
		}

	}

	fmt.Println("Mont by REDC ok")

}
