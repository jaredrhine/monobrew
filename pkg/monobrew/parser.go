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
	DefiningVar
)

const (
	commentRe = `^\s*(#|//)`
	includeRe = `(?mi)^\s*include-config\s+(\S+)$`
	newOpRe   = `(?i)^\s*new-op\s+(\S+)`
	shellRe   = `(?i)^\s*exec\s+shell\s+(from|until)\s+(\S+)`
	varRe     = `(?i)^\s*var\s+(\S+)\s+(is|until)\s+(.+)`
)

type Parser struct {
	config        *Config
	state         int
	endToken      string
	currentFile   string
	currentBlock  *Block
	currentLine   int
	currentScript string
	currentVarKey string
	currentVar    string
	inHereis      bool
}

func NewParser(config *Config) *Parser {
	return &Parser{config: config, state: TopLevel}
}

// This is an intentionally hacky, hard-coded parser. Replace as codebase and syntax stabilizes.
func (p *Parser) ParseConfig(body io.Reader) {
	v := p.verbmsg
	fileScanner := bufio.NewScanner(body)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		p.currentLine += 0
		line := fileScanner.Text()
		v(fmt.Sprint(p.state) + " processing line: " + line)

		/*
			Complete: comments at top-level in or block, new-op, exec shell from variable, exec shell until end, halt-if-fail, var variable is value, var variable until end, include-config
			In-progress:
			On-deck: state-dir
		*/

		commentPat := regexp.MustCompile(commentRe)
		newOpPat := regexp.MustCompile(newOpRe)
		shellPat := regexp.MustCompile(shellRe)
		varPat := regexp.MustCompile(varRe)

		linebytes := []byte(line)

		matchDirective := func(linebytes []byte) (matches bool) {
			return (newOpPat.Match(linebytes) || varPat.Match(linebytes))
		}

		// Handle comments by immediately skipping to next line
		if !p.inHereis && commentPat.Match(linebytes) {
			v("skipping comment")
			continue
		}

		// Handle implicit new-op close. After defining a hereis within a block, we might not notice
		// the block definition has ended until we see one of these cues. If we see one,
		// we bump ourselves back to TopLevel parsing.
		if !p.inHereis && matchDirective(linebytes) {
			v("spotted valid directive, ending block if needed and resetting to TopLevel")
			p.StoreBlock()
			p.state = TopLevel
		}

		if p.state == TopLevel {
			v("In TopLevel")
			if newOpPat.Match(linebytes) {
				v("TopLevel: starting new block")
				p.StartBlock(linebytes)
				continue

			} else if strings.HasPrefix(line, "state-dir") {
				v("TopLevel: defining state-dir TODO")
				// TODO: set state-dir
				continue

			} else if varPat.Match(linebytes) {
				v("TopLevel: found var definition")
				m := varPat.FindSubmatch(linebytes)
				variable := string(m[1])
				dir := string(m[2])
				value := string(m[3])
				if dir == "is" {
					v(fmt.Sprintf("TopLevel: setting status %s=%s", variable, value))
					p.config.SetStatusKey(variable, value)
				} else if dir == "until" {
					v(fmt.Sprintf("TopLevel: starting var until for variable=%s", variable))
					p.state = DefiningVar
					p.inHereis = true
					p.currentVar = ""
					p.currentVarKey = variable
					p.endToken = value
				}
				continue
			}
		}

		if p.state == DefiningBlock {
			v("In DefiningBlock: " + string(linebytes))
			if shellPat.Match(linebytes) {
				v("DefiningBlock: found start of exec shell")
				p.currentBlock.Touched = true
				p.currentBlock.Command = "sh"
				p.currentBlock.Args = []string{"-sex"}
				parts := regexp.MustCompile(`\s+`).Split(line, 4) // TODO: switch to shellPat reuse

				mode := strings.ToLower(parts[2])
				if mode == "until" {
					v("DefiningBlock: this is a 'exec shell until' block")
					p.state = DefiningScript
					p.inHereis = true
					p.currentScript = ""
					p.endToken = parts[3]
				} else if mode == "from" {
					v("DefiningBlock: this is a 'exec shell from' block")
					variable := parts[3]
					value := p.config.GetStatusKey(variable)
					v(fmt.Sprintf("DefiningBlock: got status key=%s val=%s", variable, value))
					p.currentBlock.Stdin = value
				}
				continue

			} else if strings.HasPrefix(line, "halt-if-fail") { // TODO: regexp for case insensitive
				p.currentBlock.HaltIfFail = true
				p.currentBlock.Touched = true
				continue
			} else {
				v("Unknown line in DefiningBlock")
				continue
			}
		}

		if p.state == DefiningScript {
			v("In DefiningScript")
			if line == p.endToken {
				v("DefiningScript: found end")
				p.currentBlock.Stdin = p.currentScript
				p.state = DefiningBlock
				p.inHereis = false
				p.currentBlock.Touched = true
				continue
			} else {
				v("DefiningScript: appending script")
				p.currentScript = p.currentScript + line + "\n"
				continue
			}
		}

		if p.state == DefiningVar {
			v("In DefiningVar")
			if line == p.endToken {
				v("DefiningVar: found end")
				p.config.SetStatusKey(p.currentVarKey, p.currentVar)
				p.currentVarKey = ""
				p.currentVar = ""
				p.state = DefiningBlock
				p.inHereis = false
				continue
			} else {
				v("DefiningVar: appending var")
				p.currentVar = p.currentVar + line + "\n"
				continue
			}
		}
	}

	p.StoreBlock()
}

func (p *Parser) StartBlock(linebytes []byte) {
	p.verbmsg("StartBlock: start")
	label := string(regexp.MustCompile(newOpRe).FindSubmatch(linebytes)[1])
	p.state = DefiningBlock
	p.StoreBlock()
	p.currentBlock = &Block{Label: label}
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
		p.currentBlock = &Block{}
	}
}

func (p *Parser) GetConfigContents(path string) string {
	pat := regexp.MustCompile(`^https?://`)
	var contents string
	if pat.Match([]byte(path)) {
		contents = p.GetUrlContents(path)
	} else {
		contents = p.GetFileContents(path)
	}
	return contents
}

func (p *Parser) GetFileContents(filepath string) string {
	readFile, err := os.Open(filepath)
	PanicIfErr(err)
	defer readFile.Close()

	bodystr, err := io.ReadAll(readFile)
	PanicIfErr(err)
	return string(bodystr)
}

func (p *Parser) GetUrlContents(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		PanicMsg("can't fetch url: " + url)
	}

	if resp.StatusCode != 200 {
		WarnMsg("Did not get an HTTP 200 back from " + url)
	}

	bodystr, err := io.ReadAll(resp.Body)
	PanicIfErr(err)
	return string(bodystr)
}

func (p *Parser) ParseFilepath(filepath string) {
	readFile, err := os.Open(filepath)
	PanicIfErr(err)
	defer readFile.Close()

	p.currentFile = filepath
	p.ParseConfig(readFile)
}

func (p *Parser) ExpandConfigs() {
	p.config.ConfigExpanded = ""

	// Start with the config paths specified on the command line
	for _, cfg := range p.config.ConfigFiles {
		p.AppendConfigToExpanded(cfg)
	}

	includePat := regexp.MustCompile(includeRe)

	includesFound := true
	loops := 0

	// Keep looping
	for includesFound {
		// But exit if we hit an infinite loop of expansion
		loops += 1
		if loops > 1000 {
			PanicMsg("exceeded include depth, probably have circular includes")
		}

		// Given "include-config foo", replace the whole string with the contents of "foo"
		replaceFunc := func(in []byte) (out []byte) {
			trimmedIn := []byte(strings.TrimSpace(string(in)))
			parts := includePat.FindSubmatch(trimmedIn)
			path := string(parts[1])
			replacement := "\n# Expanded from: " + string(trimmedIn) + "\n\n"
			replacement += p.GetConfigContents(path)
			return []byte(replacement)
		}

		p.config.ConfigExpanded = string(includePat.ReplaceAllFunc([]byte(p.config.ConfigExpanded), replaceFunc))

		if !includePat.Match([]byte(p.config.ConfigExpanded)) {
			includesFound = false
		}
	}
}

func (p *Parser) AppendConfigToExpanded(configpath string) {
	contents := p.GetConfigContents(configpath)
	p.config.ConfigExpanded += contents
}

// Logging and controlled exit

func (p *Parser) verbmsg(msg string) {
	if p.config.PrintDebug {
		fmt.Println(msg)
	}
}
