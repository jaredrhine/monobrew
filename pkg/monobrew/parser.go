package monobrew

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	TopLevel int = iota
	DefiningBlock
	DefiningScript
	DefiningStdin
)

type Parser struct {
	config        *Config
	state         int
	endToken      string
	currentFile   string
	currentBlock  *Block
	currentLine   int
	currentScript string
}

func NewParser(config *Config) *Parser {
	return &Parser{config: config, state: TopLevel}
}

// This is an intentionally hacky, hard-coded parser. Replace as codebase stabilizes.
func (p *Parser) ParseConfig(body io.Reader) {
	v := p.verbmsg
	fileScanner := bufio.NewScanner(body)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		p.currentLine += 0
		line := fileScanner.Text()
		v("processing line: " + line)

		/*
			Complete: state-dir, comments at top-level in or block, new-op
			In-progress: halt-if-fail
			On-deck: include-config
		*/
		if strings.HasPrefix(line, "#") && (p.state == TopLevel || p.state == DefiningBlock) {
			v("skipping comment")

		} else if p.state == TopLevel {
			if strings.HasPrefix(line, "new-op") {
				v("TopLevel: starting new block")
				p.StartBlock(line)

			} else if strings.HasPrefix(line, "state-dir") {
				v("TopLevel: defining state-dir TODO")
				// TODO: set state-dir
			}

		} else if p.state == DefiningBlock {
			if strings.HasPrefix(line, "new-op") {
				v("DefiningBlock: was defining block, found start of new block")
				p.StartBlock(line)
			} else if strings.HasPrefix(line, "exec shell until") { // TODO: fixup with regexp
				v("DefiningBlock: found start of exec")
				parts := regexp.MustCompile(`\s+`).Split(line, 4)
				p.currentBlock.Command = "sh"
				p.currentBlock.Args = []string{"-sex"}
				p.state = DefiningScript
				p.currentScript = ""
				p.endToken = parts[3]
			} else if strings.HasPrefix(line, "halt-if-fail") {
				p.currentBlock.HaltIfFail = true
			}

		} else if p.state == DefiningScript {
			if line == p.endToken {
				v("DefiningScript: found end")
				p.currentBlock.Stdin = p.currentScript
				p.state = DefiningBlock
			} else {
				v("DefiningScript: appending script")
				p.currentScript = p.currentScript + line + "\n"
			}
		}
	}
	p.StoreBlock()
}

func (p *Parser) ParseFile(file string) {
	pat := regexp.MustCompile(`^https?://`)
	if pat.Match([]byte(file)) {
		p.ParseUrl(file)
	} else {
		p.ParseFilepath(file)
	}
}

func (p *Parser) ParseUrl(url string) {
	resp, err := http.Get(url)
	if err != nil {
		p.parseError("can't fetch url: " + url)
	}

	if resp.StatusCode != 200 {
		p.warnmsg("Did not get an HTTP 200 back from " + url)
	}

	p.currentFile = url
	p.ParseConfig(resp.Body)
}

func (p *Parser) ParseFilepath(filepath string) {
	readFile, err := os.Open(filepath)
	PanicIfErr(err)
	defer readFile.Close()

	p.currentFile = filepath
	p.ParseConfig(readFile)
}

func (p *Parser) StartBlock(line string) {
	p.verbmsg("StartBlock: start")
	parts := regexp.MustCompile(`\s+`).Split(line, 3)
	if len(parts) < 2 {
		p.parseError("incorrect number of parts on 'new-op' line")
	}

	p.state = DefiningBlock
	p.StoreBlock()
	p.currentBlock = &Block{Touched: true, Label: parts[1]}
}

func (p *Parser) StoreBlock() {
	// We'll only store blocks if they have been modified
	p.verbmsg("StoreBlock: checking if currentBlock has been touched")
	if p.currentBlock != nil && p.currentBlock.Touched {

		// Make single-line scripts not have a trailing newline
		if strings.Count(p.currentBlock.Stdin, "\n") == 1 {
			p.currentBlock.Stdin = strings.TrimRight(p.currentBlock.Stdin, "\n")
		}

		// Add the presumed-completed block to the config
		p.verbmsg(fmt.Sprintf("StoreBlock: storing previous block - %#v", p.currentBlock))
		p.config.Blocks = append(p.config.Blocks, p.currentBlock)
	}
}

func (p *Parser) verbmsg(msg string) {
	if p.config.PrintDebug {
		fmt.Println(msg)
	}
}

func (p *Parser) warnmsg(msg string) {
	fmt.Printf("WARNING - %s\n", msg)
}

func (p *Parser) parseError(msg string) {
	msg = fmt.Sprintf("EXITING - parse error on file %s line %d - %s\n", p.currentFile, p.currentLine, msg)
	panic(msg)
}
