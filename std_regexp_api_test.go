package regexp2

import (
	"bytes"
	"fmt"
	"testing"
)

func TestStdRegexp(t *testing.T) {
	pattern := `(?sm:^[ \t]*(?<IDENT>var|const)[ \t]+(?<NAME>[a-zA-Z_][[:word:]]+)[ \t]*=[ \t]*(?<VALUE>[1-9]\d*)[ \t]*$)`
	toMatch := []string{
		`const fooBar123_ = 10`,
		`var fooBar456_ = 09`,
	}
	exp := MustCompileStd(pattern)

	var expect = [][]string{}
	var got [][]string
	for _, v := range toMatch {
		var agot []string
		agot = append(agot, fmt.Sprintf("match:%t", exp.MatchString(v)))

		got = append(got, agot)
	}
	if err := testCheckMatches(got, expect); err != nil {
		fmt.Println(testShowStringList(got))
		t.Error(err)
	}

}

func testCheckMatches(got, expect [][]string) error {
	if len(got) != len(expect) {
		return fmt.Errorf("match num mismatch, expect %d, got %d", len(expect), len(got))
	}
	for i, v := range got {
		w := expect[i]
		if len(v) != len(w) {
			return fmt.Errorf("match %d submatch-num mismatch, expect %d, got %d", i+1, len(w), len(v))
		}

		for j, vv := range v {
			if vv != w[j] {
				return fmt.Errorf(`match %d,%d submatch mismatch, expect "%s", got "%s"`, i+1, j+1, w[j], vv)
			}
		}
	}
	return nil
}

func testShowStringList(ss [][]string) string {
	var b bytes.Buffer
	b.WriteString("[][]string{\n")
	for _, v := range ss {
		b.WriteString(fmt.Sprintf("  %#v,\n", v))
	}
	b.WriteString("}")
	return b.String()
}
