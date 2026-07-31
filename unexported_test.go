package stripcomments

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnexportedDocs(test *testing.T) {
	Convey("Given source with docs on unexported declarations", test, func() {
		source := `// Package invoice builds invoices.
package invoice

// helper computes the total.
func helper() int {
	// body note
	total := 1
	return total
}

// state holds the cursor.
type state struct {
	// cursor position
	cursor int
}

// limit is the maximum.
const limit = 10
`

		Convey("By default the unexported docs survive", func() {
			out, err := StripComments([]byte(source))
			So(err, ShouldBeNil)

			got := string(out)
			So(got, ShouldContainSubstring, "// helper computes the total.")
			So(got, ShouldContainSubstring, "// state holds the cursor.")
			So(got, ShouldContainSubstring, "// cursor position")
			So(got, ShouldContainSubstring, "// limit is the maximum.")

			Convey("But body comments are still stripped", func() {
				So(got, ShouldNotContainSubstring, "body note")
			})
		})

		Convey("With DropUnexported the unexported docs are stripped", func() {
			out, err := StripComments([]byte(source), DropUnexported())
			So(err, ShouldBeNil)

			got := string(out)
			So(got, ShouldNotContainSubstring, "computes the total")
			So(got, ShouldNotContainSubstring, "holds the cursor")
			So(got, ShouldNotContainSubstring, "cursor position")
			So(got, ShouldNotContainSubstring, "is the maximum")

			Convey("And the package doc still survives", func() {
				So(got, ShouldContainSubstring, "// Package invoice builds invoices.")
			})
		})

		Convey("With DropUnexported the exported docs survive", func() {
			exportedSource := `package invoice

// Charge runs the payment.
func Charge() {}

// helper computes the total.
func helper() {}
`
			out, err := StripComments([]byte(exportedSource), DropUnexported())
			So(err, ShouldBeNil)

			got := string(out)
			So(got, ShouldContainSubstring, "// Charge runs the payment.")
			So(got, ShouldNotContainSubstring, "computes the total")
		})

		Convey("With DropExported the unexported docs survive", func() {
			out, err := StripComments([]byte(source), DropExported())
			So(err, ShouldBeNil)

			So(string(out), ShouldContainSubstring, "// helper computes the total.")
		})

		Convey("Directives survive regardless", func() {
			directiveSource := `package invoice

//go:generate stringer -type=kind

// kind is a code.
type kind int
`
			out, err := StripComments([]byte(directiveSource), DropUnexported())
			So(err, ShouldBeNil)

			got := string(out)
			So(got, ShouldContainSubstring, "//go:generate stringer -type=kind")
			So(strings.Contains(got, "is a code"), ShouldBeFalse)
		})
	})
}
