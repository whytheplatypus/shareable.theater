package main

import (
	"bytes"
	"testing"
)

func TestTemplate(t *testing.T) {
	buf := &bytes.Buffer{}
	tmpl.Execute(buf, "")
	bufTwo := &bytes.Buffer{}
	tmpl.Execute(bufTwo, "")
	if buf.String() == bufTwo.String() {
		t.Error(buf, bufTwo)
	}

}
