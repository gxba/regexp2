// adapt APIs to standard regexp

package regexp2

import (
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
	"unsafe"
)

// Bitmap used by func special to check whether a character needs to be escaped.
var specialBytes [16]byte

// special reports whether byte b needs to be escaped by QuoteMeta.
func special(b byte) bool {
	return b < utf8.RuneSelf && specialBytes[b%16]&(1<<(b/16)) != 0
}

func init() {
	for _, b := range []byte(`\.+*?()|[]{}^$`) {
		specialBytes[b%16] |= 1 << (b / 16)
	}
}

// CompileStd compile an regegular expression with standard API
func CompileStd(expr string) (*RegexpStd, error) {
	re, err := Compile(expr, RE2)
	return re.RegexpStd(), err
}

// CompileStd compile an regegular expression with standard API, it panic if expr is valid
func MustCompileStd(expr string) *RegexpStd {
	re, err := CompileStd(expr)
	if err != nil {
		panic(err)
	}
	return re
}

// RegexpStd is compiled regegular expression with standard regexp APIs export
type RegexpStd struct {
	p *Regexp
}

// RegexpStd return an compiled regegular expression with standard API
func (re *Regexp) RegexpStd() *RegexpStd {
	return &RegexpStd{p: re}
}

// RegexpStd return an compiled regegular expression with regexp2 API
func (re *RegexpStd) Regexp2() *Regexp {
	return re.p
}

// String returns the source text used to compile the regular expression.
func (re *RegexpStd) String() string {
	return re.p.String()
}

// Copy returns a new StdRegexp object copied from re.
// Calling Longest on one copy does not affect another.
//
// Deprecated: In earlier releases, when using a StdRegexp in multiple goroutines,
// giving each goroutine its own copy helped to avoid lock contention.
// As of Go 1.12, using Copy is no longer necessary to avoid lock contention.
// Copy may still be appropriate if the reason for its use is to make
// two copies with different Longest settings.
func (re *RegexpStd) Copy() *RegexpStd {
	re2 := *re.p
	return &RegexpStd{p: &re2}
}

// Longest makes future searches prefer the leftmost-longest match.
// That is, when matching against text, the regexp returns a match that
// begins as early as possible in the input (leftmost), and among those
// it chooses a match that is as long as possible.
// This method modifies the StdRegexp and may not be called concurrently
// with any other methods.
func (re *RegexpStd) Longest() {
	//TODO:
	panic("Longest unsupport")
}

// SubexpNames returns the names of the parenthesized subexpressions
// in this StdRegexp. The name for the first sub-expression is names[1],
// so that if m is a match slice, the name for m[i] is SubexpNames()[i].
// Since the StdRegexp as a whole cannot be named, names[0] is always
// the empty string. The slice should not be modified.
func (re *RegexpStd) SubexpNames() []string {
	return re.p.GetGroupNames()
}

// SubexpIndex returns the index of the first subexpression with the given name,
// or -1 if there is no subexpression with that name.
//
// Note that multiple subexpressions can be written using the same name, as in
// (?P<bob>a+)(?P<bob>b+), which declares two subexpressions named "bob".
// In this case, SubexpIndex returns the index of the leftmost such subexpression
// in the regular expression.
func (re *RegexpStd) SubexpIndex(name string) int {
	return re.p.GroupNumberFromName(name)
}

// LiteralPrefix returns a literal string that must begin any match
// of the regular expression re. It returns the boolean true if the
// literal string comprises the entire regular expression.
func (re *RegexpStd) LiteralPrefix() (prefix string, complete bool) {
	if p := re.p.code.FcPrefix; p != nil {
		return string(p.PrefixStr), p.CaseInsensitive
	} else {
		return "", false
	}
}

// MatchReader reports whether the text returned by the RuneReader
// contains any match of the regular expression re.
func (re *RegexpStd) MatchReader(r io.RuneReader) bool {
	panic("unsupport MatchReader")
}

// MatchString reports whether the string s
// contains any match of the regular expression re.
func (re *RegexpStd) MatchString(s string) bool {
	if ok, err := re.p.MatchString(s); err == nil {
		return ok
	}
	return false
}

// Match reports whether the byte slice b
// contains any match of the regular expression re.
func (re *RegexpStd) Match(b []byte) bool {
	return re.MatchString(unsafeBytesString(b))
}

// MatchString reports whether the string s
// contains any match of the regular expression pattern.
// More complicated queries need to use Compile and the full StdRegexp interface.
func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := CompileStd(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

// Match reports whether the byte slice b
// contains any match of the regular expression pattern.
// More complicated queries need to use Compile and the full StdRegexp interface.
func MatchStd(pattern string, b []byte) (matched bool, err error) {
	re, err := CompileStd(pattern)
	if err != nil {
		return false, err
	}
	return re.Match(b), nil
}

// ReplaceAllString returns a copy of src, replacing matches of the StdRegexp
// with the replacement string repl. Inside repl, $ signs are interpreted as
// in Expand, so for instance $1 represents the text of the first submatch.
func (re *RegexpStd) ReplaceAllString(src, repl string) string {
	rep, err := re.p.Replace(src, repl, 0, -1)
	if err != nil {
		fmt.Println(err)
		return src
	}
	return rep
}

// ReplaceAllLiteralString returns a copy of src, replacing matches of the StdRegexp
// with the replacement string repl. The replacement repl is substituted directly,
// without using Expand.
func (re *RegexpStd) ReplaceAllLiteralString(src, repl string) string {
	r, err := re.p.Replace(src, repl, 0, -1)
	if err != nil {
		println(err)
		return src
	}
	return r
}

// ReplaceAllStringFunc returns a copy of src in which all matches of the
// StdRegexp have been replaced by the return value of function repl applied
// to the matched substring. The replacement returned by repl is substituted
// directly, without using Expand.
func (re *RegexpStd) ReplaceAllStringFunc(src string, repl func(string) string) string {
	rep, err := re.p.ReplaceFunc(src, makeRepFunc(repl), 0, -1)
	if err != nil {
		fmt.Println(err)
		return src
	}
	return rep
}

// ReplaceAll returns a copy of src, replacing matches of the StdRegexp
// with the replacement text repl. Inside repl, $ signs are interpreted as
// in Expand, so for instance $1 represents the text of the first submatch.
func (re *RegexpStd) ReplaceAll(src, repl []byte) []byte {
	r := re.ReplaceAllString(unsafeBytesString(src), unsafeBytesString(repl))
	return []byte(r)
}

// ReplaceAllLiteral returns a copy of src, replacing matches of the StdRegexp
// with the replacement bytes repl. The replacement repl is substituted directly,
// without using Expand.
func (re *RegexpStd) ReplaceAllLiteral(src, repl []byte) []byte {
	// return re.replaceAll(src, "", 2, func(dst []byte, match []int) []byte {
	// 	return append(dst, repl...)
	// })
	panic("")
}

// ReplaceAllFunc returns a copy of src in which all matches of the
// StdRegexp have been replaced by the return value of function repl applied
// to the matched byte slice. The replacement returned by repl is substituted
// directly, without using Expand.
func (re *RegexpStd) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	// return re.replaceAll(src, "", 2, func(dst []byte, match []int) []byte {
	// 	return append(dst, repl(src[match[0]:match[1]])...)
	// })
	panic("")
}

// QuoteMeta returns a string that escapes all regular expression metacharacters
// inside the argument text; the returned string is a regular expression matching
// the literal text.
func QuoteMeta(s string) string {
	// A byte loop is correct because all metacharacters are ASCII.
	var i int
	for i = 0; i < len(s); i++ {
		if special(s[i]) {
			break
		}
	}
	// No meta characters found, so return original string.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)
	copy(b, s[:i])
	j := i
	for ; i < len(s); i++ {
		if special(s[i]) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[:j])
}

// Find returns a slice holding the text of the leftmost match in b of the regular expression.
// A return value of nil indicates no match.
func (re *RegexpStd) Find(b []byte) []byte {
	s := re.FindString(unsafeBytesString(b))
	return []byte(s)
}

// FindIndex returns a two-element slice of integers defining the location of
// the leftmost match in b of the regular expression. The match itself is at
// b[loc[0]:loc[1]].
// A return value of nil indicates no match.
func (re *RegexpStd) FindIndex(b []byte) (loc []int) {
	m, err := re.p.FindStringMatch(unsafeBytesString(b))
	if err != nil {
		println(err)
		return nil
	}
	if m != nil {
		return []int{m.Capture.Index, m.Capture.Length}
	}
	return nil
}

// FindString returns a string holding the text of the leftmost match in s of the regular
// expression. If there is no match, the return value is an empty string,
// but it will also be empty if the regular expression successfully matches
// an empty string. Use FindStringIndex or FindStringSubmatch if it is
// necessary to distinguish these cases.
func (re *RegexpStd) FindString(s string) string {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err)
		return ""
	}
	if m != nil {
		return m.Capture.String()
	}
	return ""
}

// FindStringIndex returns a two-element slice of integers defining the
// location of the leftmost match in s of the regular expression. The match
// itself is at s[loc[0]:loc[1]].
// A return value of nil indicates no match.
func (re *RegexpStd) FindStringIndex(s string) (loc []int) {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err)
		return nil
	}
	if m != nil {
		return []int{m.Capture.Index, m.Capture.Length}
	}
	return nil
}

// FindReaderIndex returns a two-element slice of integers defining the
// location of the leftmost match of the regular expression in text read from
// the RuneReader. The match text was found in the input stream at
// byte offset loc[0] through loc[1]-1.
// A return value of nil indicates no match.
func (re *RegexpStd) FindReaderIndex(r io.RuneReader) (loc []int) {
	// a := re.doExecute(r, nil, "", 0, 2, nil)
	// if a == nil {
	// 	return nil
	// }
	// return a[0:2]
	panic("unsupport FindReaderIndex")
}

// FindSubmatch returns a slice of slices holding the text of the leftmost
// match of the regular expression in b and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' descriptions in the package
// comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindSubmatch(b []byte) [][]byte {
	indexes := re.FindSubmatchIndex(b)
	result := make([][]byte, 0, len(indexes)/2)
	for i := 0; i < len(indexes)/2; i++ {
		result = append(result, b[indexes[i]:indexes[i+1]:indexes[i+1]])
	}
	return result
}

// Expand appends template to dst and returns the result; during the
// append, Expand replaces variables in the template with corresponding
// matches drawn from src. The match slice should have been returned by
// FindSubmatchIndex.
//
// In the template, a variable is denoted by a substring of the form
// $name or ${name}, where name is a non-empty sequence of letters,
// digits, and underscores. A purely numeric name like $1 refers to
// the submatch with the corresponding index; other names refer to
// capturing parentheses named with the (?P<name>...) syntax. A
// reference to an out of range or unmatched index or a name that is not
// present in the regular expression is replaced with an empty slice.
//
// In the $name form, name is taken to be as long as possible: $1x is
// equivalent to ${1x}, not ${1}x, and, $10 is equivalent to ${10}, not ${1}0.
//
// To insert a literal $ in the output, use $$ in the template.
func (re *RegexpStd) Expand(dst []byte, template []byte, src []byte, match []int) []byte {
	//return re.expand(dst, string(template), src, "", match)
	panic("")
}

// ExpandString is like Expand but the template and source are strings.
// It appends to and returns a byte slice in order to give the calling
// code control over allocation.
func (re *RegexpStd) ExpandString(dst []byte, template string, src string, match []int) []byte {
	//return re.expand(dst, template, nil, src, match)
	panic("")
}

// FindSubmatchIndex returns a slice holding the index pairs identifying the
// leftmost match of the regular expression in b and the matches, if any, of
// its subexpressions, as defined by the 'Submatch' and 'Index' descriptions
// in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindSubmatchIndex(b []byte) []int {
	m, err := re.p.FindStringMatch(unsafeBytesString(b))
	if err != nil {
		println(err.Error())
		return nil
	}

	if m != nil {
		m.populateOtherGroups()
		subs := make([]int, 0, len(m.otherGroups)+1)
		subs = append(subs, m.Group.Index)
		for i := 0; i < len(m.otherGroups); i++ {
			subs = append(subs, (&m.otherGroups[i]).Index)
		}
		return subs
	}
	return nil
}

// FindStringSubmatch returns a slice of strings holding the text of the
// leftmost match of the regular expression in s and the matches, if any, of
// its subexpressions, as defined by the 'Submatch' description in the
// package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindStringSubmatch(s string) []string {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	if m != nil {
		m.populateOtherGroups()
		subs := make([]string, 0, len(m.otherGroups)+1)
		subs = append(subs, m.Group.String())
		for i := 0; i < len(m.otherGroups); i++ {
			subs = append(subs, (&m.otherGroups[i]).String())
		}
		return subs
	}
	return nil
}

// FindStringSubmatchIndex returns a slice holding the index pairs
// identifying the leftmost match of the regular expression in s and the
// matches, if any, of its subexpressions, as defined by the 'Submatch' and
// 'Index' descriptions in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindStringSubmatchIndex(s string) []int {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	if m != nil {
		m.populateOtherGroups()
		subs := make([]int, 0, 2*(len(m.otherGroups)+1))
		subs = append(subs, m.Group.Index, m.Group.Length)
		for i := 0; i < len(m.otherGroups); i++ {
			g := &m.otherGroups[i]
			subs = append(subs, g.Index, g.Length)
		}
		return subs
	}
	return nil
}

// FindReaderSubmatchIndex returns a slice holding the index pairs
// identifying the leftmost match of the regular expression of text read by
// the RuneReader, and the matches, if any, of its subexpressions, as defined
// by the 'Submatch' and 'Index' descriptions in the package comment. A
// return value of nil indicates no match.
func (re *RegexpStd) FindReaderSubmatchIndex(r io.RuneReader) []int {
	//return re.pad(re.doExecute(r, nil, "", 0, re.prog.NumCap, nil))
	panic("unsupport FindReaderSubmatchIndex")
}

// FindAll is the 'All' version of Find; it returns a slice of all successive
// matches of the expression, as defined by the 'All' description in the
// package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAll(b []byte, n int) [][]byte {
	m, err := re.p.FindStringMatch(unsafeBytesString(b))
	if err != nil {
		println(err.Error())
		return nil
	}

	var result [][]byte
	for m != nil {
		result = append(result, b[m.Group.Index:m.Group.Length:m.Group.Length])

		m, err = re.p.FindNextMatch(m)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return result
}

// FindAllIndex is the 'All' version of FindIndex; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description
// in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllIndex(b []byte, n int) [][]int {
	return re.FindAllStringIndex(unsafeBytesString(b), n)
}

// FindAllString is the 'All' version of FindString; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description
// in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllString(s string, n int) []string {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	var result []string
	for m != nil {
		result = append(result, m.Group.String())

		m, err = re.p.FindNextMatch(m)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return result
}

// FindAllStringIndex is the 'All' version of FindStringIndex; it returns a
// slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllStringIndex(s string, n int) [][]int {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	var result [][]int
	for m != nil {
		result = append(result, []int{m.Group.Index, m.Group.Length})

		m, err = re.p.FindNextMatch(m)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return result
}

// FindAllSubmatch is the 'All' version of FindSubmatch; it returns a slice
// of all successive matches of the expression, as defined by the 'All'
// description in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllSubmatch(b []byte, n int) [][][]byte {
	if n < 0 {
		n = len(b) + 1
	}
	matches := re.FindAllSubmatchIndex(b, n)
	var result = make([][][]byte, 0, len(matches))
	for _, m := range matches {
		matchSize := len(m) / 2
		match := make([][]byte, 0, matchSize)
		for j := 0; j < matchSize; j++ {
			sub := b[m[2*j]:m[2*j+1]:m[2*j+1]]
			match = append(match, sub)
		}
		result = append(result, match)
	}

	return result
}

// FindAllSubmatchIndex is the 'All' version of FindSubmatchIndex; it returns
// a slice of all successive matches of the expression, as defined by the
// 'All' description in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllSubmatchIndex(b []byte, n int) [][]int {
	s := unsafeBytesString(b)
	return re.FindAllStringSubmatchIndex(s, n)
}

// FindAllStringSubmatch is the 'All' version of FindStringSubmatch; it
// returns a slice of all successive matches of the expression, as defined by
// the 'All' description in the package comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllStringSubmatch(s string, n int) [][]string {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	var result [][]string
	for m != nil {
		m.populateOtherGroups()
		subs := make([]string, 0, len(m.otherGroups)+1)
		subs = append(subs, m.Group.String())
		for i := 0; i < len(m.otherGroups); i++ {
			subs = append(subs, (&m.otherGroups[i]).String())
		}
		result = append(result, subs)

		m, err = re.p.FindNextMatch(m)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return result
}

// FindAllStringSubmatchIndex is the 'All' version of
// FindStringSubmatchIndex; it returns a slice of all successive matches of
// the expression, as defined by the 'All' description in the package
// comment.
// A return value of nil indicates no match.
func (re *RegexpStd) FindAllStringSubmatchIndex(s string, n int) [][]int {
	m, err := re.p.FindStringMatch(s)
	if err != nil {
		println(err.Error())
		return nil
	}

	var result [][]int
	for m != nil {
		m.populateOtherGroups()
		subs := make([]int, 0, len(m.otherGroups)+1)
		subs = append(subs, m.Group.Index)
		for i := 0; i < len(m.otherGroups); i++ {
			subs = append(subs, (&m.otherGroups[i]).Index)
		}
		result = append(result, subs)

		m, err = re.p.FindNextMatch(m)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return result
}

// Split slices s into substrings separated by the expression and returns a slice of
// the substrings between those expression matches.
//
// The slice returned by this method consists of all the substrings of s
// not contained in the slice returned by FindAllString. When called on an expression
// that contains no metacharacters, it is equivalent to strings.SplitN.
//
// Example:
//   s := regexp.MustCompile("a*").Split("abaabaccadaaae", 5)
//   // s: ["", "b", "b", "c", "cadaaae"]
//
// The count determines the number of substrings to return:
//   n > 0: at most n substrings; the last substring will be the unsplit remainder.
//   n == 0: the result is nil (zero substrings)
//   n < 0: all substrings
func (re *RegexpStd) Split(s string, n int) []string {
	if n == 0 {
		return nil
	}

	if len(re.p.pattern) > 0 && len(s) == 0 {
		return []string{""}
	}

	matches := re.FindAllStringIndex(s, n)
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]
	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}

// makeRepFunc convert a standard replace function to regexp2 replace function
func makeRepFunc(f func(string) string) MatchEvaluator {
	return func(m Match) string {
		return f(m.String())
	}
}

// unsafeStringBytes return GoString's buffer slice(enable modify string)
func unsafeStringBytes(s string) []byte {
	var bh reflect.SliceHeader
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// unsafeBytesString convert b to string without copy
func unsafeBytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
