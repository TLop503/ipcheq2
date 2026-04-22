package data
package data

import "testing"

func TestEmbeddedDataContainsExpectedFile(t *testing.T) {
	f, err := embeddedData.Open("cyberghost.txt")
	if err != nil {
		t.Fatalf("expected embedded file cyberghost.txt to be present: %v", err)
	}
	defer f.Close()
}
