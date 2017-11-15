package parser

import (
	"bytes"
	"log"

	"github.com/wavefronthq/go-proxy/common"
)

const MAX_BUFFER_SIZE = 2

// Parser represents a parser.
type PointParser struct {
	s   *PointScanner
	buf struct {
		tok []Token  // last read n tokens
		lit []string // last read n literals
		n   int      // unscanned buffer size (max=2)
	}
	scanBuf  bytes.Buffer // buffer reused for scanning tokens
	writeBuf bytes.Buffer // buffer reused for parsing elements
	Elements []ElementParser
}

// Returns a slice of ElementParser's for the Graphite format
func NewGraphiteElements() []ElementParser {
	var elements []ElementParser
	wsParser := WhiteSpaceParser{}
	repeatParser := LoopedParser{wrappedParser: &TagParser{}, wsPaser: &wsParser}
	elements = append(elements, &NameParser{}, &wsParser, &ValueParser{}, &wsParser,
		&TimestampParser{optional: true}, &wsParser, &repeatParser)
	return elements
}

// Returns a slice of ElementParser's for the OpenTSDB format
func NewOpenTSDBElements() []ElementParser {
	var elements []ElementParser
	wsParser := WhiteSpaceParser{}
	repeatParser := LoopedParser{wrappedParser: &TagParser{}, wsPaser: &wsParser}
	elements = append(elements, &LiteralParser{literal: "put"}, &wsParser, &NameParser{}, &wsParser,
		&TimestampParser{}, &wsParser, &ValueParser{}, &wsParser, &repeatParser)
	return elements
}

// Returns new instance of Graphite format specific parser
func NewGraphiteParser() *PointParser {
	elements := NewGraphiteElements()
	return &PointParser{Elements: elements}
}

func NewOpenTSDBParser() *PointParser {
	elements := NewOpenTSDBElements()
	return &PointParser{Elements: elements}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that from the internal buffer instead.
func (p *PointParser) scan() (Token, string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		idx := p.buf.n % MAX_BUFFER_SIZE
		tok, lit := p.buf.tok[idx], p.buf.lit[idx]
		p.buf.n -= 1
		return tok, lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buffer(tok, lit)

	return tok, lit
}

func (p *PointParser) buffer(tok Token, lit string) {
	// create the buffer if it is empty
	if len(p.buf.tok) == 0 {
		p.buf.tok = make([]Token, MAX_BUFFER_SIZE)
		p.buf.lit = make([]string, MAX_BUFFER_SIZE)
	}

	// for now assume a simple circular buffer of length two
	p.buf.tok[0], p.buf.lit[0] = p.buf.tok[1], p.buf.lit[1]
	p.buf.tok[1], p.buf.lit[1] = tok, lit
}

// unscan pushes the previously read token back onto the buffer.
func (p *PointParser) unscan() {
	p.unscanTokens(1)
}

func (p *PointParser) unscanTokens(n int) {
	if n > MAX_BUFFER_SIZE {
		// just log for now
		log.Printf("cannot unscan more than %d tokens", MAX_BUFFER_SIZE)
	}
	p.buf.n += n
}

func (p *PointParser) reset(b []byte) {

	// reset the scan buffer and write new byte
	p.scanBuf.Reset()
	p.scanBuf.Write(b)

	if p.s == nil {
		p.s = NewScanner(&p.scanBuf)
	} else {
		// reset p.s.r passing in the buffer as the reader
		p.s.r.Reset(&p.scanBuf)
	}
	p.buf.n = 0
}

// Parses one entire pointLine
func (p *PointParser) Parse(b []byte) (*common.Point, error) {
	p.reset(b)
	point := common.Point{}
	for _, element := range p.Elements {
		err := element.parse(p, &point)
		if err != nil {
			return nil, err
		}
	}
	return &point, nil
}
