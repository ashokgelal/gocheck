package gocheck_test

import (
	"errors"
	"launchpad.net/gocheck"
	"reflect"
	"runtime"
)

type CheckersS struct{}

var _ = gocheck.Suite(&CheckersS{})

func testInfo(c *gocheck.C, checker gocheck.Checker, name string, paramNames []string) {
	info := checker.Info()
	if info.Name != name {
		c.Fatalf("Got name %s, expected %s", info.Name, name)
	}
	if !reflect.DeepEqual(info.Params, paramNames) {
		c.Fatalf("Got param names %#v, expected %#v", info.Params, paramNames)
	}
}

func testCheck(c *gocheck.C, checker gocheck.Checker, result bool, error string, params ...interface{}) ([]interface{}, []string) {
	info := checker.Info()
	names := append([]string{}, info.Params...)
	result_, error_ := checker.Check(params, names)
	if result_ != result || error_ != error {
		c.Fatalf("%s.Check(%#v) returned (%#v, %#v) rather than (%#v, %#v)",
			info.Name, params, result_, error_, result, error)
	}
	return params, names
}

func (s *CheckersS) TestBug(c *gocheck.C) {
	bug := gocheck.Bug("a %d bc", 42)
	info := bug.GetBugInfo()
	if info != "a 42 bc" {
		c.Fatalf("Bug() returned %#v", info)
	}
}

func (s *CheckersS) TestIsNil(c *gocheck.C) {
	testInfo(c, gocheck.IsNil, "IsNil", []string{"value"})

	testCheck(c, gocheck.IsNil, true, "", nil, nil, true, "")
	testCheck(c, gocheck.IsNil, false, "", "a", nil, false, "")

	testCheck(c, gocheck.IsNil, true, "", (chan int)(nil), nil)
	testCheck(c, gocheck.IsNil, false, "", make(chan int), nil)
	testCheck(c, gocheck.IsNil, true, "", (error)(nil), nil)
	testCheck(c, gocheck.IsNil, false, "", errors.New(""), nil)
	testCheck(c, gocheck.IsNil, true, "", ([]int)(nil), nil)
	testCheck(c, gocheck.IsNil, false, "", make([]int, 1), nil)
	testCheck(c, gocheck.IsNil, false, "", int(0), nil)
}

func (s *CheckersS) TestNotNil(c *gocheck.C) {
	testInfo(c, gocheck.NotNil, "NotNil", []string{"value"})

	testCheck(c, gocheck.NotNil, false, "", nil, nil)
	testCheck(c, gocheck.NotNil, true, "", "a", nil)

	testCheck(c, gocheck.NotNil, false, "", (chan int)(nil))
	testCheck(c, gocheck.NotNil, true, "", make(chan int))
	testCheck(c, gocheck.NotNil, false, "", (error)(nil))
	testCheck(c, gocheck.NotNil, true, "", errors.New(""))
	testCheck(c, gocheck.NotNil, false, "", ([]int)(nil))
	testCheck(c, gocheck.NotNil, true, "", make([]int, 1))
}

func (s *CheckersS) TestNot(c *gocheck.C) {
	testInfo(c, gocheck.Not(gocheck.IsNil), "Not(IsNil)", []string{"value"})

	testCheck(c, gocheck.Not(gocheck.IsNil), false, "", nil)
	testCheck(c, gocheck.Not(gocheck.IsNil), true, "", "a")
}

type simpleStruct struct {
	i int
}

func (s *CheckersS) TestEquals(c *gocheck.C) {
	testInfo(c, gocheck.Equals, "Equals", []string{"obtained", "expected"})

	// The simplest.
	testCheck(c, gocheck.Equals, true, "", 42, 42)
	testCheck(c, gocheck.Equals, false, "", 42, 43)

	// Different native types.
	testCheck(c, gocheck.Equals, false, "", int32(42), int64(42))

	// With nil.
	testCheck(c, gocheck.Equals, false, "", 42, nil)

	// Arrays
	testCheck(c, gocheck.Equals, true, "", []byte{1, 2}, []byte{1, 2})
	testCheck(c, gocheck.Equals, false, "", []byte{1, 2}, []byte{1, 3})

	// Struct values
	testCheck(c, gocheck.Equals, true, "", simpleStruct{1}, simpleStruct{1})
	testCheck(c, gocheck.Equals, false, "", simpleStruct{1}, simpleStruct{2})

	// Struct pointers
	testCheck(c, gocheck.Equals, true, "", &simpleStruct{1}, &simpleStruct{1})
	testCheck(c, gocheck.Equals, false, "", &simpleStruct{1}, &simpleStruct{2})
}

func (s *CheckersS) TestMatches(c *gocheck.C) {
	testInfo(c, gocheck.Matches, "Matches", []string{"value", "regex"})

	// Simple matching
	testCheck(c, gocheck.Matches, true, "", "abc", "abc")
	testCheck(c, gocheck.Matches, true, "", "abc", "a.c")

	// Must match fully
	testCheck(c, gocheck.Matches, false, "", "abc", "ab")
	testCheck(c, gocheck.Matches, false, "", "abc", "bc")

	// String()-enabled values accepted
	testCheck(c, gocheck.Matches, true, "", errors.New("abc"), "a.c")
	testCheck(c, gocheck.Matches, false, "", errors.New("abc"), "a.d")

	// Some error conditions.
	testCheck(c, gocheck.Matches, false, "Obtained value is not a string and has no .String()", 1, "a.c")
	testCheck(c, gocheck.Matches, false, "Can't compile regex: error parsing regexp: missing closing ]: `[c$`", "abc", "a[c")
}

func (s *CheckersS) TestPanics(c *gocheck.C) {
	testInfo(c, gocheck.Panics, "Panics", []string{"function", "expected"})

	// Plain strings.
	testCheck(c, gocheck.Panics, true, "", func() { panic("BOOM") }, "BOOM", true, "")
	testCheck(c, gocheck.Panics, false, "", func() { panic("KABOOM") }, "BOOM", false, "")
	testCheck(c, gocheck.Panics, true, "", func() bool { panic("BOOM") }, "BOOM", true, "")

	// Error values.
	testCheck(c, gocheck.Panics, true, "", func() { panic(errors.New("BOOM")) }, errors.New("BOOM"), true, "")
	testCheck(c, gocheck.Panics, false, "", func() { panic(errors.New("KABOOM")) }, errors.New("BOOM"), false, "")

	// String matching.
	testCheck(c, gocheck.Panics, true, "", func() { panic(errors.New("BOOM")) }, "BO.M", true, "")
	testCheck(c, gocheck.Panics, false, "", func() { panic(errors.New("KABOOM")) }, "BO.M", false, "")

	// Some errors.
	testCheck(c, gocheck.Panics, false, "Function has not panicked", func() bool { return false }, "BOOM")
	testCheck(c, gocheck.Panics, false, "Function must take zero arguments", 1, "BOOM")

	// Verify params/names mutation
	params, names := testCheck(c, gocheck.Panics, false, "", func() { panic(errors.New("KABOOM")) }, errors.New("BOOM"), false, "")
	c.Assert(params[0], gocheck.Equals, errors.New("KABOOM"))
	c.Assert(names[0], gocheck.Equals, "panic")
}

func (s *CheckersS) TestFitsTypeOf(c *gocheck.C) {
	testInfo(c, gocheck.FitsTypeOf, "FitsTypeOf", []string{"obtained", "sample"})

	// Basic types
	testCheck(c, gocheck.FitsTypeOf, true, "", 1, 0)
	testCheck(c, gocheck.FitsTypeOf, false, "", 1, int64(0))

	// Aliases
	testCheck(c, gocheck.FitsTypeOf, false, "", 1, errors.New(""))
	testCheck(c, gocheck.FitsTypeOf, false, "", "error", errors.New(""))
	testCheck(c, gocheck.FitsTypeOf, true, "", errors.New("error"), errors.New(""))

	// Structures
	testCheck(c, gocheck.FitsTypeOf, false, "", 1, simpleStruct{})
	testCheck(c, gocheck.FitsTypeOf, false, "", simpleStruct{42}, &simpleStruct{})
	testCheck(c, gocheck.FitsTypeOf, true, "", simpleStruct{42}, simpleStruct{})
	testCheck(c, gocheck.FitsTypeOf, true, "", &simpleStruct{42}, &simpleStruct{})

	// Some bad values
	testCheck(c, gocheck.FitsTypeOf, false, "Invalid sample value", 1, interface{}(nil))
	testCheck(c, gocheck.FitsTypeOf, false, "", interface{}(nil), 0)
}

func (s *CheckersS) TestImplements(c *gocheck.C) {
	testInfo(c, gocheck.Implements, "Implements", []string{"obtained", "ifaceptr"})

	var e error
	var re runtime.Error
	testCheck(c, gocheck.Implements, true, "", errors.New(""), &e)
	testCheck(c, gocheck.Implements, false, "", errors.New(""), &re)

	// Some bad values
	testCheck(c, gocheck.Implements, false, "ifaceptr should be a pointer to an interface variable", 0, errors.New(""))
	testCheck(c, gocheck.Implements, false, "ifaceptr should be a pointer to an interface variable", 0, interface{}(nil))
	testCheck(c, gocheck.Implements, false, "", interface{}(nil), &e)
}
