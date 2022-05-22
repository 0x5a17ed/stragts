package stragts

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos pos      // The starting position, in bytes, of this item in the input string.
	val string   // The simpleValue of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemNil
	itemBool
	itemNumber
	itemString
	itemIdentifier
	itemEnable
	itemDisable
	itemAssign
	itemListSeparator
	itemArgumentSeparator
)

const eof = -1

type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input string    // the string being scanned
	pos   pos       // current position in the input
	start pos       // start position of this item
	atEOF bool      // we have hit the end of input and returned eof
	items chan item // channel of scanned items
}

// run executes the state machine for the lexer.
func (l *lexer) run() {
	for state := lexArgumentStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// next returns and consumes the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.atEOF = true
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += pos(w)
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.undo()
	return r
}

// undo steps back one rune.
func (l *lexer) undo() {
	if !l.atEOF && l.pos > 0 {
		_, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= pos(w)
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.undo()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.undo()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.item.
func (l *lexer) errorf(format string, args ...any) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// item returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) item() item {
	return <-l.items
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces. Also catches cases
// like "$x+2" not being acceptable without a space, in case we decide one
// day to implement arithmetic.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, ',', ';', '=':
		return true
	}
	return false
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{input: input, items: make(chan item)}
	go l.run()
	return l
}

// lexArgumentStart scans a single argument field.
func lexArgumentStart(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case r == '~' || r == '!':
		if !unicode.IsLetter(l.peek()) {
			return l.errorf("bad character %#U", r)
		}
		if r == '!' {
			l.emit(itemDisable)
		} else {
			l.emit(itemEnable)
		}
		return lexIdentifier
	case unicode.IsLetter(r):
		l.undo()
		return lexIdentifier
	case isNumeric(r):
		l.undo()
		return lexNumber
	case r == '"' || r == '\'':
		l.undo()
		return lexQuote
	default:
		return l.errorf("bad character %#U", r)
	}
}

// lexInArgument scans a single argument field.
func lexInArgument(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case r == ',':
		l.emit(itemArgumentSeparator)
		return lexArgumentStart
	case r == ';':
		l.emit(itemListSeparator)
		return lexValue
	case r == '=':
		l.emit(itemAssign)
		return lexValue
	default:
		return l.errorf("bad character %#U", r)
	}
}

// lexValue scans a single string, integer, or identifier simpleValue.
func lexValue(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		return l.errorf("assignment missing simpleValue")
	case unicode.IsLetter(r):
		l.undo()
		return lexIdentifier
	case isNumeric(r):
		l.undo()
		return lexNumber
	case r == '"' || r == '\'':
		l.undo()
		return lexQuote
	default:
		return l.errorf("bad character %#U", r)
	}
}

// lexIdentifier scans a single identifier.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case r == '-':
			fallthrough
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.undo()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch word {
			case "true", "false":
				l.emit(itemBool)
			case "nil":
				l.emit(itemNil)
			default:
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexInArgument
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
	closingQuote := l.next()
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof {
				break
			}
			fallthrough
		case eof:
			return l.errorf("unterminated quoted string")
		case closingQuote:
			break Loop
		}
	}
	l.emit(itemString)
	return lexInArgument
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	if !l.atTerminator() {
		return l.errorf("bad character %#U", l.peek())
	}
	l.emit(itemNumber)
	return lexInArgument
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789_"
	if l.accept("0") {
		// Note: Leading 0 does not mean octal in floats.
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF_"
		} else if l.accept("oO") {
			digits = "01234567_"
		} else if l.accept("bB") {
			digits = "01_"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if len(digits) == 10+1 && l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	if len(digits) == 16+6+1 && l.accept("pP") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	return true
}
