package stripcomments

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefault_PreservesPackageDoc(t *testing.T) {
	Convey("Given source with a package doc and an unexported func doc", t, func() {
		src := `// Package bank handles payments.
package bank

// helper does math.
func helper() {}
`
		Convey("When stripped with default options", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the package doc survives", func() {
				So(got, ShouldContainSubstring, "// Package bank handles payments.")
			})
			Convey("Then the unexported func doc survives", func() {
				So(got, ShouldContainSubstring, "// helper does math.")
			})
		})
	})
}

func TestDefault_PreservesExportedFuncDoc(t *testing.T) {
	Convey("Given source with an exported func doc and a body comment", t, func() {
		src := `package p

// Charge runs the payment.
func Charge() {
	x := 1 // temp
	_ = x
}
`
		Convey("When stripped with default options", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the exported func doc survives", func() {
				So(got, ShouldContainSubstring, "// Charge runs the payment.")
			})
			Convey("Then the body comment is stripped", func() {
				So(got, ShouldNotContainSubstring, "temp")
			})
		})
	})
}

func TestDefault_PreservesExportedTypeAndFieldDoc(t *testing.T) {
	Convey("Given source with type, exported field, and unexported field docs", t, func() {
		src := `package p

// Config holds settings.
type Config struct {
	// Timeout in seconds.
	Timeout int
	// internal cursor
	pos int
}
`
		Convey("When stripped with default options", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the exported type doc survives", func() {
				So(got, ShouldContainSubstring, "// Config holds settings.")
			})
			Convey("Then the exported field doc survives", func() {
				So(got, ShouldContainSubstring, "// Timeout in seconds.")
			})
			Convey("Then the unexported field doc survives by default", func() {
				So(got, ShouldContainSubstring, "// internal cursor")
			})
		})
	})
}

func TestDropExported_StripsExportedFuncDoc(t *testing.T) {
	Convey("Given source with an exported func doc", t, func() {
		src := `package p

// Charge runs the payment.
func Charge() {}
`
		Convey("When stripped with DropExported", func() {
			out, err := StripComments([]byte(src), DropExported())
			So(err, ShouldBeNil)

			Convey("Then the exported func doc is stripped", func() {
				So(string(out), ShouldNotContainSubstring, "Charge runs the payment")
			})
		})
	})
}

func TestDropPackage_StripsPackageDoc(t *testing.T) {
	Convey("Given source with a package doc", t, func() {
		src := `// Package bank handles payments.
package bank
`
		Convey("When stripped with DropPackage", func() {
			out, err := StripComments([]byte(src), DropPackage())
			So(err, ShouldBeNil)

			Convey("Then the package doc is stripped", func() {
				So(string(out), ShouldNotContainSubstring, "handles payments")
			})
		})
	})
}
