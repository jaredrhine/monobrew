package monobrew

import (
	"encoding/json"
	"os"
	"time"

	"golang.org/x/sys/unix"
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

	PrintVerboseResult  bool
	PrintDebug          bool
	NukeStateDirAtStart bool

	Parser  *Parser
	Scanner *Scanner
}

func NewConfig() *Config {
	config := &Config{
		StateDir:           "/var/tmp/monobrew",
		PrintVerboseResult: false,
		Blocks:             make([]*Block, 0),
		Status:             make(map[string]string, 0),
	}
	config.Parser = NewParser(config)
	config.Scanner = NewScanner(config)
	return config
}

func (c *Config) OrderedOps() []*Block {
	return c.Blocks
}

func (c *Config) EnsureEnv() {
	if c.NukeStateDirAtStart {
		err := os.RemoveAll(c.StateDir)
		if err != nil {
			panic("problems nuking directory " + c.StateDir)
		}
	}

	if unix.Access(c.StateDir, unix.W_OK) != nil {
		os.MkdirAll(c.StateDir, os.ModePerm)
	}
}

func (c *Config) Load() {
	c.Scanner.Scan()

	for _, cfg := range c.ConfigFiles {
		c.Parser.ParseFile(cfg)
	}

	c.EnsureEnv()
}

func (c *Config) AddConfigFile(file string) {
	c.ConfigFiles = append(c.ConfigFiles, file)
}

func (c *Config) AddCmdBlock(name string, command string) {
	b := CmdBlock(name, command)
	c.Blocks = append(c.Blocks, b)
}

func (c *Config) GetStatusKey(key string) string {
	return c.Status[key]
}

func (c *Config) SetStatusKey(key string, value string) {
	c.Status[key] = value
}

func CmdBlock(label string, command string) *Block {
	block := Block{Label: label, Stdin: command, Command: "sh", Args: []string{"-sex"}, Success: false}
	return &block
}
