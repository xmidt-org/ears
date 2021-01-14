package bit_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/bit"
)

const (
	b1 bit.Mask = 1 << iota
	b2
	b3
	b4
)

var bArr = []bit.Mask{b1, b2, b3, b4}

func TestMaskSet(t *testing.T) {

	for _, v := range bArr {
		t.Run(v.String(), func(t *testing.T) {

			a := NewWithT(t)
			var b bit.Mask
			a.Expect(b).To(Equal(bit.Mask(0)))
			b.Set(v)
			a.Expect(b == v).To(BeTrue())

		})
	}
}

func TestMaskClear(t *testing.T) {
	for _, v := range bArr {
		t.Run(v.String(), func(t *testing.T) {

			a := NewWithT(t)
			var b bit.Mask
			b.Set(v)
			a.Expect(b == v).To(BeTrue())
			b.Clear(v)
			a.Expect(b).To(Equal(bit.Mask(0)))

		})
	}

}

func TestMaskIsSet(t *testing.T) {
	for _, v := range bArr {
		t.Run(v.String(), func(t *testing.T) {
			a := NewWithT(t)
			var b bit.Mask
			b.Set(v)
			a.Expect(b.IsSet(v)).To(BeTrue())
			b.Clear(v)
			a.Expect(b.IsSet(v)).To(BeFalse())

		})
	}
}

func TestMaskFlip(t *testing.T) {
	for _, v := range bArr {
		t.Run(v.String(), func(t *testing.T) {
			a := NewWithT(t)
			var b bit.Mask
			b.Flip(v)
			a.Expect(b.IsSet(v)).To(BeTrue())
			a.Expect(b == v).To(BeTrue())

			b.Flip(v)
			a.Expect(b.IsSet(v)).To(BeFalse())
			a.Expect(b != v).To(BeTrue())
		})
	}

}

func TestMaskString(t *testing.T) {
	g := goldie.New(t, goldie.WithTestNameForDir(true))

	for i, v := range bArr {
		g.Assert(t, fmt.Sprintf("%d", i), []byte(v.String()))
	}
}
