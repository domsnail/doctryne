package javascript_parsers

import (
	"context"
	"io"
)

type Parser struct {
	ctx context.Context

	file     io.Reader
	lockfile io.Reader
}

func (p *Parser) WithContext(ctx context.Context) *Parser {
	p.ctx = ctx
	return p
}

func (p *Parser) WithFile(file io.Reader) *Parser {
	p.file = file
	return p
}

func (p *Parser) WithLockfile(lockfile io.Reader) *Parser {
	p.lockfile = lockfile
	return p
}
