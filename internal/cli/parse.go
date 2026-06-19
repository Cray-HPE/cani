/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package cli

import (
	"fmt"
	"strings"
)

// parse interprets args as a mix of flags and positional arguments, setting
// flags on the set and returning the positionals.  It reports helpRequested
// when -h/--help is encountered so the caller can short-circuit to help.
func (f *FlagSet) parse(args []string) (positionals []string, helpRequested bool, err error) {
	i := 0
	for i < len(args) {
		tok := args[i]
		switch {
		case tok == "--":
			positionals = append(positionals, args[i+1:]...)
			return positionals, false, nil
		case strings.HasPrefix(tok, "--"):
			consumed, help, perr := f.parseLong(tok[2:], args, i)
			if perr != nil {
				return nil, false, perr
			}
			if help {
				return positionals, true, nil
			}
			i += consumed
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			consumed, help, perr := f.parseShort(tok[1:], args, i)
			if perr != nil {
				return nil, false, perr
			}
			if help {
				return positionals, true, nil
			}
			i += consumed
		default:
			positionals = append(positionals, tok)
			i++
		}
	}
	return positionals, false, nil
}

// parseLong handles a "--name" or "--name=value" token and returns how many
// argv tokens it consumed.
func (f *FlagSet) parseLong(body string, args []string, i int) (consumed int, help bool, err error) {
	name, value, hasValue := splitOnEquals(body)
	if name == "help" && f.Lookup("help") == nil {
		return 1, true, nil
	}
	flag := f.Lookup(name)
	if flag == nil {
		return 0, false, fmt.Errorf("unknown flag: --%s", name)
	}
	if hasValue {
		return 1, false, f.set(flag, value)
	}
	if flag.NoOptDefVal != "" {
		return 1, false, f.set(flag, flag.NoOptDefVal)
	}
	if i+1 >= len(args) {
		return 0, false, fmt.Errorf("flag needs an argument: --%s", name)
	}
	return 2, false, f.set(flag, args[i+1])
}

// parseShort handles a "-abc" shorthand cluster and returns how many argv
// tokens it consumed.  A non-bool flag consumes the remainder of the cluster
// (or the next token) as its value.
func (f *FlagSet) parseShort(body string, args []string, i int) (consumed int, help bool, err error) {
	for j := 0; j < len(body); j++ {
		ch := body[j : j+1]
		if ch == "h" && f.shorthandLookup("h") == nil {
			return 1, true, nil
		}
		flag := f.shorthandLookup(ch)
		if flag == nil {
			return 0, false, fmt.Errorf("unknown shorthand flag: %q in -%s", ch, body)
		}
		if flag.NoOptDefVal != "" {
			if err = f.set(flag, flag.NoOptDefVal); err != nil {
				return 0, false, err
			}
			continue
		}
		return f.shortValue(flag, body[j+1:], args, i)
	}
	return 1, false, nil
}

// shortValue assigns a non-bool shorthand flag its value, taken either from the
// remainder of the cluster ("-ovalue" / "-o=value") or the next argv token.
func (f *FlagSet) shortValue(flag *Flag, rest string, args []string, i int) (consumed int, help bool, err error) {
	if len(rest) > 0 {
		rest = strings.TrimPrefix(rest, "=")
		return 1, false, f.set(flag, rest)
	}
	if i+1 >= len(args) {
		return 0, false, fmt.Errorf("flag needs an argument: -%s", flag.Shorthand)
	}
	return 2, false, f.set(flag, args[i+1])
}

// splitOnEquals splits "name=value" into its parts, reporting whether a value
// was present.
func splitOnEquals(body string) (name, value string, hasValue bool) {
	if eq := strings.IndexByte(body, '='); eq >= 0 {
		return body[:eq], body[eq+1:], true
	}
	return body, "", false
}
