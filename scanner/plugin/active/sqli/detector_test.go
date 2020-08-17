package sqli_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/plugin/active/sqli"
)

func TestDetector(t *testing.T) {
	falseTest := "<html><body>You made it!</body></html>"
	expectedTest := "<html><body>You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near ''\"' at line 1</body></html>"
	d := sqli.NewDetector()
	result, matched := d.Detect([]byte(falseTest))
	if result != browserk.Unknown {
		t.Fatalf("matched on false test %s", matched)
	}

	result, matched = d.Detect([]byte(expectedTest))
	if result != browserk.DBMySQL {
		t.Fatalf("did not match DBMySQL %s", matched)
	}
}
