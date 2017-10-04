package parser

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/wavefronthq/go-proxy/common"
)

var (
	ErrEOF = errors.New("EOF")
)

// Interface for parsing line elements.
type ElementParser interface {
	parse(p *PointParser, pt *common.Point) error
}

type NameParser struct{}
type ValueParser struct{}
type TimestampParser struct{}
type WhiteSpaceParser struct{}
type TagParser struct{}
type LoopedParser struct {
	wrappedParser ElementParser
	wsPaser       *WhiteSpaceParser
}

func (ep *NameParser) parse(p *PointParser, pt *common.Point) error {
	//Valid characters are: a-z, A-Z, 0-9, hyphen ("-"), underscore ("_"), dot (".").
	// Forward slash ("/") and comma (",") are allowed if metricName is enclosed in double quotes.
	name, err := parseLiteral(p)
	if err != nil {
		return err
	}
	pt.Name = name
	return nil
}

func (ep *ValueParser) parse(p *PointParser, pt *common.Point) error {
	tok, lit := p.scan()
	if tok == EOF {
		return fmt.Errorf("found %q, expected number", lit)
	}

	p.writeBuf.Reset()
	if tok == MINUS_SIGN {
		p.writeBuf.WriteString(lit)
		tok, lit = p.scan()
	}

	for tok != EOF && (tok == LETTER || tok == NUMBER || tok == DOT) {
		p.writeBuf.WriteString(lit)
		tok, lit = p.scan()
	}
	p.unscan()

	pt.Value = p.writeBuf.String()
	_, err := strconv.ParseFloat(pt.Value, 64)
	if err != nil {
		return fmt.Errorf("invalid metric value %s", pt.Value)
	}
	return nil
}

func (ep *TimestampParser) parse(p *PointParser, pt *common.Point) error {
	tok, lit := p.scan()
	if tok == EOF {
		return fmt.Errorf("found %q, expected number", lit)
	}

	if tok != NUMBER {
		p.unscanTokens(2)
		return setTimestamp(pt, 0, 1)
	}

	p.writeBuf.Reset()
	for tok != EOF && tok == NUMBER {
		p.writeBuf.WriteString(lit)
		tok, lit = p.scan()
	}
	p.unscan()

	tsStr := p.writeBuf.String()
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return err
	}
	return setTimestamp(pt, ts, len(tsStr))
}

func setTimestamp(pt *common.Point, ts int64, numDigits int) error {

	if numDigits == 19 {
		// nanoseconds
		ts = ts / 1e9
	} else if numDigits == 16 {
		// microseconds
		ts = ts / 1e6
	} else if numDigits == 13 {
		// milliseconds
		ts = ts / 1e3
	}

	if ts == 0 {
		ts = getCurrentTime()
	}
	pt.Timestamp = ts
	return nil
}

func (ep *LoopedParser) parse(p *PointParser, pt *common.Point) error {
	for {
		err := ep.wrappedParser.parse(p, pt)
		if err != nil {
			return err
		}
		err = ep.wsPaser.parse(p, pt)
		if err == ErrEOF {
			break
		}
	}
	return nil
}

func (ep *TagParser) parse(p *PointParser, pt *common.Point) error {
	k, err := parseLiteral(p)
	if err != nil {
		if k == "" {
			return nil
		}
		return err
	}

	next, lit := p.scan()
	if next != EQUALS {
		return fmt.Errorf("found %q, expected equals", lit)
	}

	v, err := parseLiteral(p)
	if err != nil {
		return err
	}
	if len(pt.Tags) == 0 {
		pt.Tags = make(map[string]string)
	}
	pt.Tags[k] = v
	return nil
}

func (ep *WhiteSpaceParser) parse(p *PointParser, pt *common.Point) error {
	tok, lit := p.scan()
	if tok != WS {
		if tok == EOF {
			return ErrEOF
		}
		return fmt.Errorf("found %q, expected whitespace", lit)
	}
	return nil
}

func parseQuotedLiteral(p *PointParser) (string, error) {
	p.writeBuf.Reset()

	//TODO: handle escaped quote scenario
	tok, lit := p.scan()
	for tok != EOF && tok != QUOTES {
		// let everything through
		p.writeBuf.WriteString(lit)
		tok, lit = p.scan()
	}
	if tok == EOF {
		return "", fmt.Errorf("found %q, expected quotes", lit)
	}
	return p.writeBuf.String(), nil
}

func parseLiteral(p *PointParser) (string, error) {
	tok, lit := p.scan()
	if tok == EOF {
		return "", fmt.Errorf("found %q, expected literal", lit)
	}

	if tok == QUOTES {
		return parseQuotedLiteral(p)
	}

	//TODO: handle quotes
	p.writeBuf.Reset()
	for tok != EOF && tok > literal_beg && tok < literal_end {
		p.writeBuf.WriteString(lit)
		tok, lit = p.scan()
	}
	p.unscan()
	return p.writeBuf.String(), nil
}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / 1e9
}
