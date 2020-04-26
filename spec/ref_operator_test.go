package spec

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestRefOperator(t *testing.T) {
	t.Run("ref-operator", func(t *testing.T) {
		v, _ := ParseRefOperator("sys/xxxx:1.2.3")

		NewWithT(t).Expect(v.Version.Major()).To(Equal(uint64(1)))
		NewWithT(t).Expect(v.Version.Minor()).To(Equal(uint64(2)))
		NewWithT(t).Expect(v.Version.Patch()).To(Equal(uint64(3)))

		NewWithT(t).Expect(v.String()).To(Equal("sys/xxxx:1.2.3"))
	})
}
