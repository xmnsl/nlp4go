//
// Package strings implements unicode string abstraction
// that operates with runes (symbols), not bytes
//
// Basic usage:
//   ustr := NewString("this is unicode string: こんにちは")
//   slen := ustr.Length()
//
//   var str string
//   str = ustr.String() // equal to original string
//   usub := ustr.Substring(8, 15) // cut `unicode` word
//
//   re, _ := regexp.Compile("string:.*")
//   matched := ustr.Match(re)  // matched == true
//   loc := ustr.FindFirst(re)
//   usub = ustr.Substring(loc[0], loc[1])  // cut `string: こんにちは`
//
//   re, _ = regexp.Compile("i")
//   locs := ustr.FindAll(re)  // locs contain 4 locations (one for every "i" symbol in string)
//
//   newS, cnt := ustr.Replace(re, "Ё")  // cnt == 4, newS.String() == "thЁs Ёs unЁcode strЁng: こんにちは""
//

package core

import (
	"regexp"
)

// Error type that returns ReadRune function
// when end of string occured
type noRuneError struct {
}

// Returns error description for noRuneError
func (e noRuneError) Error() string {
	return "no next rune in string"
}

// Representation of string as an array of runes.
type String struct {
	runes []rune
	ridx  int // used for runes indexing in regexp
}

// Creates new String object from `string`
// and returns pointer to this object
func NewString(orig string) *String {
	return &String{runes: []rune(orig)}
}

// Returns length of String object (count of runes)
func (s *String) Length() int {
	return len(s.runes)
}

// Converts String object to `string`
func (s *String) String() string {
	return string(s.runes)
}

// Returns substring as a pointer to new String object
func (s *String) Substring(from, to int) *String {
	if from < 0 || from > to || to > len(s.runes) {
		return nil
	}
	runes := s.runes[from:to]
	new_s := String{runes: runes}
	return &new_s
}

// This function is a part of io.RuneReader iterface
// that used in regexp
func (s *String) ReadRune() (r rune, size int, err error) {
	if s.ridx == len(s.runes) {
		return 0, 0, noRuneError{}
	}
	r = s.runes[s.ridx]
	err = nil
	s.ridx += 1
	// padding to 4 bytes per rune
	// expecting offset returned from regexp.FindReaderIndex equal to rune_offset*4
	size = 4
	return
}

// This internal function convert bytes location to runes location in string
func (s *String) normalizeRegexpLoc(runes_offset, from, to int) (loc []int) {
	from_rune := runes_offset + (from / 4)
	to_rune := runes_offset + (to / 4)
	loc = []int{from_rune, to_rune}
	return loc
}

// Finds first substring location found by regexp
// If not found - returns `nil`
// Location is an 2 bytes array {from, to}
func (s *String) FindFirst(re *regexp.Regexp) (loc []int) {
	s.ridx = 0
	b_loc := re.FindReaderIndex(s)
	if b_loc == nil {
		s.ridx = 0
		return nil
	}
	s.ridx = 0
	return s.normalizeRegexpLoc(0, b_loc[0], b_loc[1])
}

// Finds all substrings locations in String object by regexp
// FindAll function returns array of locations (location is an 2 bytes array {from,to})
func (s *String) FindAll(re *regexp.Regexp) [][]int {
	s.ridx = 0
	var ret_locs [][]int
	var cur_ridx int
	for {
		cur_ridx = s.ridx
		b_loc := re.FindReaderIndex(s)
		if b_loc == nil {
			s.ridx = 0
			break
		} else {
			s.ridx = cur_ridx + b_loc[1]/4
		}
		ret_locs = append(ret_locs, s.normalizeRegexpLoc(cur_ridx, b_loc[0], b_loc[1]))

	}
	return ret_locs
}

// Match reports whether the regexp matches the String object
// returns `true` if matched and `false` otherwise
func (s *String) Match(re *regexp.Regexp) bool {
	s.ridx = 0
	matched := re.MatchReader(s)
	s.ridx = 0
	return matched
}

// Replace finds substrings in String object by regexp and replace it by `newText`
// Returns new String object and count of replaced strings
func (s *String) Replace(re *regexp.Regexp, newText string) (*String, int) {
	newStr := NewString(newText)
	locs := s.FindAll(re)
	cut_len := 0
	for _, loc := range locs {
		cut_len += loc[1] - loc[0]
	}
	nlen := newStr.Length()
	new_len := nlen*len(locs) + s.Length() - cut_len
	runes := make([]rune, new_len)
	s_idx := 0
	d_idx := 0
	d_end := 0
	for _, loc := range locs {
		d_end = d_idx + loc[0] - s_idx
		copy(runes[d_idx:d_end], s.runes[s_idx:loc[0]])
		copy(runes[d_end:d_end+nlen], newStr.runes)
		s_idx = loc[1]
		d_idx = d_end + nlen
	}
	copy(runes[d_idx:], s.runes[s_idx:])
	new := String{runes: runes}
	return &new, len(locs)
}
