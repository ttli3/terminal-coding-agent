## Overview
Based on skeleton code from [Thorsten Ball's "How to Build an Agent"][1]. We create a CLI for
an agentic coding experience, enhanced with tool-calling capabilities to read + list +
edit files, execute shell commands, and generate diffs.

[1]: https://ampcode.com/how-to-build-an-agent

## Installation

### Option 1: Install from source

1. Clone this repository
   ```bash
   git clone https://github.com/ttli3/terminal-coding-agent.git
   cd terminal-coding-agent
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Build the binary
   ```bash
   go build -o coding-agent
   ```

4. (Optional) Move the binary to your PATH
   ```bash
   sudo mv coding-agent /usr/local/bin/
   ```

### Option 2: Install with Go Install

```bash
go install github.com/ttli3/terminal-coding-agent@latest
```

### Option 3: Use the installation script

```bash
git clone https://github.com/ttli3/terminal-coding-agent.git
cd terminal-coding-agent
./install.sh
```

### Option 4: Use Docker

```bash
git clone https://github.com/ttli3/terminal-coding-agent.git
cd terminal-coding-agent
docker-compose up
```

Or build and run the Docker image directly:

```bash
docker build -t coding-agent .
docker run -it -e ANTHROPIC_API_KEY=your_api_key_here coding-agent
```

## Config

Bring your own Anthropic API key. You can set it up in one of two ways:

1. Create a `.env` file in the directory where you run the agent:
   ```
   ANTHROPIC_API_KEY=your_api_key_here
   ```

2. Set it as an environment variable:
   ```bash
   export ANTHROPIC_API_KEY=your_api_key_here
   ```

## Usage

If you installed the binary to your PATH:
```bash
coding-agent
```

If you're running from the source directory:
```bash
go run main.go
```

Or if you built the binary but didn't move it:
```bash
./coding-agent
```

Using the wrapper script:
```bash
./coding-agent.sh
```

Using Make:
```bash
make run
```

Once running, you can chat with the agent and run various coding tasks.

## Current Tools

- **read_file**: Read the contents of a file
- **list_files**: List files in a directory
- **edit_file**: Make changes to a file, with diff preview
- **run_command**: Execute shell commands
- **generate_diff**: Show differences between two versions of code
