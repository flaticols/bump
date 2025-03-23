# Custom Help for Cobra CLI

This document explains how to integrate and use the custom help command in your Cobra CLI application.

## Overview

The custom help implementation provides a more visually appealing and organized help output for your CLI commands. It replaces the default Cobra help command and adds the following features:

- Colorized output sections for better readability
- Better organization of command information
- Custom-styled header and sections
- Improved flag descriptions
- Better subcommand listing
- Box-styled header for your application name

## How to Integrate

1. Copy the `cmd/help.go` file into your project's command directory.
2. In your `main.go` file, create a `HelpOptions` struct and call `SetupCustomHelp` on your root command.

Example:

```go
// Create the root command
rootCmd := cmd.CreateRootCmd(options)

// Setup help options
helpOpts := &cmd.HelpOptions{
    ErrPrinter:     color.New(color.FgRed).SprintfFunc(),
    InfoPrinter:    color.New(color.FgWhite).SprintfFunc(),
    WarningPrinter: color.New(color.FgYellow).SprintfFunc(),
    OkPrinter:      color.New(color.FgGreen).SprintfFunc(),
}

// Apply custom help
cmd.SetupCustomHelp(rootCmd, helpOpts)
```

## Enhancement Tips

To make your commands more helpful, provide detailed information:

1. Set a descriptive `Long` field for each command to explain its purpose
2. Add an `Example` field with real-world usage examples
3. Use proper flag descriptions
4. Organize subcommands logically

```go
cmd.Long = `This is a detailed description of the command. 
It can span multiple lines and should thoroughly explain 
what the command does and when to use it.`

cmd.Example = `  # Example with explanation
  app command --flag value
  
  # Another example
  app command subcommand`
```

## Usage with Subcommands

The custom help system automatically applies to all subcommands. When you add a new subcommand, it will inherit the custom help format:

```go
// Create a subcommand
subCmd := &cobra.Command{
    Use:   "subcommand",
    Short: "Short description",
    Long:  "Detailed description of the subcommand",
}

// Add it to the parent
rootCmd.AddCommand(subCmd)
```

No additional setup is required for subcommands as the help system is applied recursively.

## Customization

You can modify the `help.go` file to adjust:

- The header box style and width
- Section titles
- Colors and formatting
- Layout and spacing
- Help sections organization

The most common customization point is in the `printCustomHelp` function where you can rearrange sections or add new ones.
