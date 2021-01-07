package match_test

import (
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/plugins/match"

	. "github.com/onsi/gomega"
)

// NOTE: Ideally we put in a pull request to go-enum
// to generate the test files as well to give 100% coverage
// for all genenerated code

func TestModeType(t *testing.T) {

	a := NewWithT(t)

	for _, e := range match.ModeTypeValues() {
		m, err := match.ParseModeTypeString(e.String())
		a.Expect(err).To(BeNil())
		a.Expect(m).To(Equal(e))
		a.Expect(e.Registered()).To(BeTrue())

		a.Expect(
			match.ModeTypeSliceContains(match.ModeTypeValues(), e),
		).To(BeTrue())

		a.Expect(
			match.ModeTypeSliceContainsAny(match.ModeTypeValues(), e),
		).To(BeTrue())

		// Text
		{
			b, err := e.MarshalText()
			a.Expect(b).To(Equal([]byte(e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.ModeType(-1)
			err = (&e2).UnmarshalText(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// Binary
		{
			b, err := e.MarshalBinary()
			a.Expect(b).To(Equal([]byte(e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.ModeType(-1)
			err = (&e2).UnmarshalBinary(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// JSON
		{
			b, err := e.MarshalJSON()
			a.Expect(string(b)).To(Equal(fmt.Sprintf(`"%s"`, e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.ModeType(-1)
			err = (&e2).UnmarshalJSON(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// YAML
		{
			b, err := e.MarshalYAML()
			a.Expect(b).To(Equal(e.String()))
			a.Expect(err).To(BeNil())

			e2 := match.ModeType(-1)
			err = yaml.Unmarshal([]byte(b.(string)), &e2)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

	}

}

func TestModeTypeErrors(t *testing.T) {

	a := NewWithT(t)

	// Invalid mode type value
	m := match.ModeType(23)
	a.Expect(m.String()).To(Equal(`ModeType(23)`))
	a.Expect(m.Registered()).To(BeFalse())
	a.Expect(
		match.ModeTypeSliceContains(match.ModeTypeValues(), m),
	).To(BeFalse())
	a.Expect(
		match.ModeTypeSliceContainsAny(match.ModeTypeValues(), m),
	).To(BeFalse())

	// Invalid mode type string
	m, err := match.ParseModeTypeString("invalid string goes here")
	a.Expect(m).To(Equal(match.ModeType(0)))
	a.Expect(err).ToNot(BeNil())

	// Invalid JSON unmarshal
	err = m.UnmarshalJSON([]byte("}x4"))
	a.Expect(err).ToNot(BeNil())

	// Invalid YAML unmarshal
	err = m.UnmarshalYAML(func(interface{}) error { return fmt.Errorf("error") })
	a.Expect(err).ToNot(BeNil())

}
