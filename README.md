# Monobrew: A single-binary host configuration management tool

Author: Jared Rhine &lt;jared@wordzoo.com&gt;

Last update: June 2023

## Background

I periodically set up new Linux machines (new physical hosts and new VMs), and have particular ways I like to configure those boxes.

The common config management tools (chef, puppet, ansible, salt) are based on dynamic/scripting languages and require a substantial runtime installation to function. Installing one isn't a hard task, but involves multiple steps and has side effects. Docker-related configuration management tools generally don't apply to this use case. Terraform can be configured to run locally but isn't a common mode.

My pile of bootstrapping shell scripts is functional, but I still needed to get ssh set up and check out repos which contained those scripts. Large shell scripts are a maintenance hassle and I dreamt of using a config file instead for my basic "overwrite this file" and "install this package" early bootstrapping needs.

So I found myself just wanting a simple way to download just one file (binary) and run a command pointed at a config file. As golang compiles to a single binary easily, this `monobrew` codebase to provide simple host configuration tasks was born.

## Use cases

- Setting up a newly-installed Linux host in a repeatable, documented way
- Keeping machines in sync by using the same configuration files

## Design goals

- Single binary, downloadable
- Simple config file. Able to run multiple config files. Able to fetch config from URLs provided on command line.
- Predictable ordering. Scripts are executed from top to bottom.
- Can embed multiple shell scripts inside one config file

## Status

Code maturity:

- Written by a golang beginner
- Not battle tested. Error conditions are rarely handled robustly.
- No tests
- No community feedback. This codebase is optimized for the author's specific use cases, with little effort to support wide flexibility and all common config management cases.

Works:

- Config can be local files or URLs
- Runs shell scripts
- Captures runtime info about each executed block into files, including JSON-parseable

Future directions:

- More directives to match common (chef, puppet, ansible, salt) operations and flow control options
- More robust parser. Support expectations around whitespace and comments. Report accurately on errors.

## Installation

`monobrew` is intended to be distributed as a single binary. No binary is publicly distributed at this time via github, so you should compile the package yourself using go. Please search the web for instructions on how to set up a working go compiler in your environment.

```shell
git clone https://github.com/jaredrhine/monobrew
go build -o monobrew -ldflags="-s -w" cmd/monobrew/main.go
```

## Usage

1. Create a configuration file and place it in the filesystem or on a web server
1. Download the `monobrew` binary
1. Run `monobrew` with one or more config files specified on the command line:

```shell
wget https://files.example.com/boot/monobrew
./monobrew --config https://files.example.com/boot/my-monobrew.conf
./monobrew --config https://files.example.com/boot/my-monobrew.conf --config local-changes.conf
```

## Configuration

All `monobrew` operations are driven by one or more configuration files.

### Config file example

```text
# See https://github.com/jaredrhine/monobrew for docs
# Usage: ./monobrew --config https://example.com/my.conf

state-dir /var/tmp/monobrew

new-op initial
script shell until ENDBLOCKSCRIPT
echo Hi
ENDBLOCKSCRIPT

new-op sudo-jared
script shell until end
echo "jared NOPASSWD: ALL" > /etc/sudoers.d/jared
end

# TODO: install a package
# TODO: use go templating
# TODO: use script defined once in two different ops via templating
# TODO: set environment variables
# TODO: set alternate executable
# TODO: pass in stdin
# TODO: only run once
# TODO: run as an alternate user
# TODO: set a timeout
# TODO: wait for completion of blocks executing in parallel
```

### Config file specification

- Parsing the config file results in an ordered sequence of "ops" (aka "blocks") which are executed by the `Runner` in order.
- Valid config file directives are:
  - `new-op [LABEL]`
  - `exec shell until [END]`
  - `halt-if-fail`
  - `state-dir [PATH_TO_DIR]`
- `state-dir [DIRECTORY]`
  - A `state-dir` directive will change the default `/var/tmp/monobrew` directory to use your preferred directory instead. If present, it must be specify a filesystem path to a directory. The path will be created if it doesn't already exist, and `monobrew` will record a number of files to that directory while executing the config.
- `new-op [LABEL]`
  - Define a new op using a `new-op` directive. The directives below the `new-op` line
  - Each `new-op` "resets" the block definition to an empty state; all directives specified above the `new-op` line are not persistent. Each `new-op` directive needs a unique "label" placed after `new-op`, like: `new-op delete-all-the-things`. This label is used only for reporting. No whitespace is allowed.
  - `exec shell until [END]`
    - Inside (underneath) a `new-op` directive, define a shell script to be executed for that op using `exec shell until [HEREIS]` directive.
    - All text between the "script" line and the HEREIS line is used as the body of the script. You can use embedded newlines and indentation as you'd prefer.
    - `shell` means "regular shell". The command used for this op will be `/bin/sh` (not BASH), and the `-ex` switches will be set (for "exit on any error" and "trace the script").
    - `[HEREIS]` can be any string (square brackets are not required).
  - `halt-if-fail`
    - Set `halt-if-fail` within a `new-op` directive to cause `monobrew` to exit if that op returns any exit code except 0 representing success.
- Comments and whitespace
  - Blank newlines are ignored (but maintained within `exec` tags).
  - Comments are defined by a line starting with a hash character, with optional whitespace before the hash mark.
- Each execution of a block by the `Runner` records a set of files in the directory specified by the `state-dir` directive. There are three files created:
  - `${SEQUENCE}.${BLOCK-NAME}.output` - the merged stdout + stderr from the executed op
  - `${SEQUENCE}.${BLOCK-NAME}.exitcode` - contains a single line with a single number recording the exit code of the executed op
  - `${SEQUENCE}.${BLOCK-NAME}.run` - JSON file with metadata about the inputs, outputs, and context of the executed op
    - `label` (string) - a unique name for the block, as specified on the `new-op` line
    - `opCounter` (integer) - the order or sequence is which this block was run
    - `command` (string) - the executable run for this op
    - `commandPath` (string) - the full file path to the command, after searching `PATH`
    - `args` (list of strings) - the command line parameters passed to the executable
    - `stdin` (string) - the standard input that was passed to the executable
    - `success` (boolean) - true if the executable run without errors (that is, exit code was 0)
    - `exitCode` (integer) - the exit code returned by the completed executable
    - `haltIfFail` (boolean) - true if `halt-on-fail` was configured for this op
    - `startTime` (iso8601 timestamp) - the timestamp when the executable started to run
    - `endTime` (iso8601 timestamp) - the timestamp when the executable completed
    - `elapsedTime` (float) - wall-clock time spent by the executable (end time minus start time)
    - `runError` (string) - a string recording the error if the executable was not able to be started
    - `outputIsEmpty` (boolean) - true if there was no output at all returned by the executable
    - `outputFile` (string) - the file path where the combined standard output and standard error from the executable can be found
- _(TODO)_ Whitespace before and after directives is ignored.
- _(TODO)_ Directives are case insensitive.
- _(TODO)_ You can split configuration into multiple files using the `include-config` directive.
- _(TODO)_ Each script block is executed in a separate and concurrent goroutine.
- _(TODO)_ Any executed block without a `parallel` directive causes the `Runner` to wait for the script to exit before proceeding.
- _(TODO)_ You can wait for all previous blocks to exit by placing a `wait-for-all-to-exit` directive, as many times as needed.
- _(TODO)_ An implicit `wait-for-all-to-exit` is inserted at the end of the config blocks.
- _(TODO)_ An `only-once` directive on a block will skip the block if a `${STATE-DIR}/${BLOCK-NAME}.output` exists.

## Code structure

- CLI: `cmd/monobrew/main.go`
  - Handles CLI switches and kicks off a monobrew run
- Runner: `pkg/monobrew/runner.go`
  - Executes ops in order, recording the state of each
- Config: `pkg/monobrew/config.go`
  - Stores all state for `monobrew` execution included the parsed list of ops
- Parser: `pkg/monobrew/parser.go`
  - Parses `monobrew` config file format
  - Knows how to use local files and perform HTTP fetches

## Work plan

- ~~Framework to execute blocks and record the results~~
- ~~Basic parser of config files can configure blocks~~
- ~~Record start/end/elapsed timestamps for execution of ops~~
- ~~Auto-add `set -xe` to beginning of script block~~
- ~~Can specify multiple configs~~
- ~~Config files can be URLs~~
- ~~Option to exit if command doesn't succeed~~
- ~~Make state-dir optional via default value~~
- ~~Rename "script" to "exec"~~
- include-config
- Able to set a variable. `var [VARNAME] is [VALUE]` and `var [VARNAME] until [END]`
- Run only once, tracked via state file
- Scan machine (like puppet catalog), able to detect OS
- Package installation
- Ensure a package is not present (command-not-found)
- Conditional execution for blocks. Only run block if conditional is true.
- Status data structure maintained
- Support another shell mode which actually processes stdin (rather than using stdin to pass in script to /bin/sh)
- Golang template processing
- Auto-add a trap to beginning of script block
- Able to set own shell or direct command
- Loop over and fetch configs up-front then concatenate, then parse to load definitions, then parse again to execute(?)
- git checkout
- git update
- git gc
- asdf install
- asdf refresh
- Support command timeout, using os.exec.CommandContext
- Support setting env vars
- Handle reuse of new-op label (either redefine with warning or panic)
- Record user the command was run as
- Blocks are run in goroutines
- Optionally blow away the state dir before the run
- Usage docs
- Library of built-ins TBD
- Utility function to map string to boolean (t/f/yes/no/true/false/0/1)
- Multipass config parsing. Able to define scripts below their usage.
- Sudoers setup
- Tests
- Run ops in a way to be able to monitor the output while it's running, for long running scripts
- Run as user
- RedHat heritage

## Design notes

### Desired operations

- Ensure a package is installed (apt, brew)
- Ensure a package is not installed (apt, brew)
- Ensure a file has specified contents, optionally templatized. Sourced locally or over HTTP.
- Ensure symlink is in place
- Install asdf in personal account
- Install update asdf plugins, install/update asdf configuration
- Ensure a git repo is checked out
- Ensure a git repo is up to date
- Ensure git repos are gc'ed
- Ensure docker containers are installed
- Produce text status report
- Ensure files have certain permissions
- Run a script. Locally and over ssh.
- Create user
- Ensure sudo setup

### Better handled by `flygaze` alerting/monitoring

- Report on unchecked in changes across git repos
- Alert if disk space is low
- Alert if uptime drops

### Maybe

- Scripting language
- Ability to layer code units
- Daemon mode? runit unit, systemd unit
- Install as daemon mode
- Integrate with flygaze and monitoring?
