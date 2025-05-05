## Terminal Coding Agent
Implementation of Thorsten Ball's ["How to Build an Agent"][1] tutorial. We create a CLI for
an agentic coding experience, enhanced with tool-calling capabilities to read + list +
edit files, execute shell commands, and generate diffs.

[1]: https://ampcode.com/how-to-build-an-agent


## Setup

1. Clone this repository
2. Create a `.env` file in the project root with your Anthropic API key:
   ```
   ANTHROPIC_API_KEY=your_api_key_here
   ```
3. Run `go mod download` to install dependencies

## Running the Agent

```bash
go run main.go
```

Once running, you can chat with Claude and request assistance with various tasks. The agent will execute tools as needed to fulfill your requests.

## Tool Definitions

- **read_file**: Read the contents of a file
- **list_files**: List files in a directory
- **edit_file**: Make changes to a file, with diff preview
- **run_command**: Execute shell commands
- **generate_diff**: Show differences between two versions of code

