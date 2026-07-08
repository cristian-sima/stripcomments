package stripcomments

import (
	"strings"
	"testing"
)

func TestDefault_PreservesPackageDoc(t *testing.T) {
	src := `// Package bank handles payments.
package bank

// helper does math.
func helper() {}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "// Package bank handles payments.") {
		t.Errorf("expected package doc to survive by default, got:\n%s", got)
	}
	if strings.Contains(got, "does math") {
		t.Errorf("expected unexported func doc to be stripped, got:\n%s", got)
	}
}

func TestDefault_PreservesExportedFuncDoc(t *testing.T) {
	src := `package p

// Charge runs the payment.
func Charge() {
	x := 1 // temp
	_ = x
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "// Charge runs the payment.") {
		t.Errorf("expected exported func doc to survive by default, got:\n%s", got)
	}
	if strings.Contains(got, "temp") {
		t.Errorf("expected body comment to be stripped, got:\n%s", got)
	}
}

func TestDefault_PreservesExportedTypeAndFieldDoc(t *testing.T) {
	src := `package p

// Config holds settings.
type Config struct {
	// Timeout in seconds.
	Timeout int
	// internal cursor
	pos int
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "// Config holds settings.") {
		t.Errorf("expected exported type doc to survive, got:\n%s", got)
	}
	if !strings.Contains(got, "// Timeout in seconds.") {
		t.Errorf("expected exported field doc to survive, got:\n%s", got)
	}
	if strings.Contains(got, "internal cursor") {
		t.Errorf("expected unexported field doc to be stripped, got:\n%s", got)
	}
}

func TestDropExported_StripsExportedFuncDoc(t *testing.T) {
	src := `package p

// Charge runs the payment.
func Charge() {}
`
	out, err := StripComments([]byte(src), DropExported())
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	if strings.Contains(string(out), "Charge runs the payment") {
		t.Errorf("expected exported func doc to be stripped with DropExported, got:\n%s", out)
	}
}

func TestDropPackage_StripsPackageDoc(t *testing.T) {
	src := `// Package bank handles payments.
package bank
`
	out, err := StripComments([]byte(src), DropPackage())
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	if strings.Contains(string(out), "handles payments") {
		t.Errorf("expected package doc to be stripped with DropPackage, got:\n%s", out)
	}
}
