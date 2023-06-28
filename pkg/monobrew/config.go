package monobrew

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Block struct {
	Touched          bool          `json:"-"`
	Label            string        `json:"label"`
	OpCounter        uint          `json:"opCounter"`
	Command          string        `json:"command"`
	CommandPath      string        `json:"commandPath"`
	Args             []string      `json:"args"`
	Stdin            string        `json:"stdin"`
	Success          bool          `json:"success"`
	ExitCode         int           `json:"exitCode"`
	HaltIfFail       bool          `json:"haltIfFail"`
	StartTime        time.Time     `json:"startTime"`
	EndTime          time.Time     `json:"endTime"`
	ElapsedTime      time.Duration `json:"-"`
	RunError         string        `json:"runError"`
	StdouterrIsEmpty bool          `json:"outputIsEmpty"`
	Stdouterr        string        `json:"-"`
	StdouterrFile    string        `json:"outputFile"`
}

func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block
	return json.Marshal(&struct {
		RunTime float64 `json:"runTime"`
		*Alias
	}{
		RunTime: b.ElapsedTime.Seconds(),
		Alias:   (*Alias)(b),
	})
}

type Config struct {
	StateDir    string
	ConfigFiles []string

	Blocks []*Block
	Status map[string]string

	PrintVerboseResult bool

	Parser *Parser
}

func NewConfig() *Config {
	config := &Config{
		StateDir:           "/var/tmp/monobrew",
		PrintVerboseResult: false,
		Blocks:             make([]*Block, 0),
		Status:             make(map[string]string, 0),
	}
	config.Parser = NewParser(config)
	return config
}

func (c *Config) OrderedOps() []*Block {
	return c.Blocks
}

func (c *Config) Load() {
	mkdirscript := fmt.Sprintf("mkdir -p %s; chmod 700 %s", c.StateDir, c.StateDir)
	c.AddCmdBlock("make-monobrew-statedir", mkdirscript)
}

func (c *Config) InitEnv() {
	err := os.MkdirAll(c.StateDir, 0755)
	PanicIfErr(err)
}

func (c *Config) Init() {
	c.Load()
	c.InitEnv()
	for _, cfg := range c.ConfigFiles {
		c.Parser.ParseFile(cfg)
	}
}

func (c *Config) AddConfigFile(file string) {
	c.ConfigFiles = append(c.ConfigFiles, file)
}

func (c *Config) AddCmdBlock(name string, command string) {
	b := CmdBlock(name, command)
	c.Blocks = append(c.Blocks, b)
}

func CmdBlock(label string, command string) *Block {
	block := Block{Label: label, Stdin: command, Command: "sh", Args: []string{"-sex"}, Success: false}
	return &block
}
