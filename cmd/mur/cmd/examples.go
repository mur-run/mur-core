package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Browse and install example patterns",
	Long: `Browse and install example patterns to get started.

Examples:
  mur examples              # List available examples
  mur examples install all  # Install all examples
  mur examples install go   # Install Go examples
  mur examples show <name>  # Preview an example`,
	RunE: runExamples,
}

var examplesInstallCmd = &cobra.Command{
	Use:   "install <category|all>",
	Short: "Install example patterns",
	Args:  cobra.ExactArgs(1),
	RunE:  runExamplesInstall,
}

var examplesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Preview an example pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  runExamplesShow,
}

func init() {
	rootCmd.AddCommand(examplesCmd)
	examplesCmd.AddCommand(examplesInstallCmd)
	examplesCmd.AddCommand(examplesShowCmd)
}

type examplePattern struct {
	name        string
	category    string
	description string
	content     string
}

var builtinExamples = []examplePattern{
	// Go patterns
	{
		name:        "go-error-handling",
		category:    "go",
		description: "Go error handling best practices",
		content: `id: go-error-handling
name: go-error-handling
description: Go error handling best practices

content: |
  When writing Go code:
  - Always check errors immediately after function calls
  - Use fmt.Errorf with %w to wrap errors with context
  - Create custom error types for domain-specific errors
  - Use errors.Is() and errors.As() for error checking
  - Return early on errors to reduce nesting
  
  Example:
    if err != nil {
        return fmt.Errorf("failed to process %s: %w", name, err)
    }

tags:
  confirmed: [go, error-handling, best-practices]
schema_version: 2
`,
	},
	{
		name:        "go-testing",
		category:    "go",
		description: "Go testing patterns and table-driven tests",
		content: `id: go-testing
name: go-testing
description: Go testing patterns and table-driven tests

content: |
  When writing Go tests:
  - Use table-driven tests for multiple cases
  - Name test cases descriptively
  - Use t.Helper() in helper functions
  - Use t.Parallel() for independent tests
  - Put tests in *_test.go files
  
  Example table-driven test:
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"empty", "", "", true},
        {"valid", "hello", "HELLO", false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ToUpper(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }

tags:
  confirmed: [go, testing, best-practices]
schema_version: 2
`,
	},
	{
		name:        "go-concurrency",
		category:    "go",
		description: "Go concurrency patterns with goroutines and channels",
		content: `id: go-concurrency
name: go-concurrency
description: Go concurrency patterns with goroutines and channels

content: |
  Go concurrency best practices:
  - Use channels for communication between goroutines
  - Use sync.WaitGroup to wait for goroutines to complete
  - Use context.Context for cancellation and timeouts
  - Avoid sharing memory; share by communicating
  - Use sync.Mutex only when necessary
  
  Common patterns:
  - Worker pool: fixed goroutines processing jobs from channel
  - Fan-out/fan-in: distribute work, collect results
  - Pipeline: chain of processing stages
  
  Example with context:
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    select {
    case result := <-ch:
        return result, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }

tags:
  confirmed: [go, concurrency, goroutines, channels]
schema_version: 2
`,
	},

	// Swift patterns
	{
		name:        "swift-error-handling",
		category:    "swift",
		description: "Swift error handling with throws and Result",
		content: `id: swift-error-handling
name: swift-error-handling
description: Swift error handling with throws and Result

content: |
  Swift error handling best practices:
  - Define errors as enums conforming to Error
  - Use throws for synchronous operations
  - Use Result<T, Error> for async callbacks
  - Use async throws for modern async code
  - Provide descriptive error messages
  
  Example:
    enum ValidationError: LocalizedError {
        case empty
        case tooShort(min: Int)
        
        var errorDescription: String? {
            switch self {
            case .empty: return "Value cannot be empty"
            case .tooShort(let min): return "Must be at least \(min) characters"
            }
        }
    }
    
    func validate(_ input: String) throws -> String {
        guard !input.isEmpty else { throw ValidationError.empty }
        guard input.count >= 3 else { throw ValidationError.tooShort(min: 3) }
        return input
    }

tags:
  confirmed: [swift, error-handling, best-practices]
schema_version: 2
`,
	},
	{
		name:        "swift-async-await",
		category:    "swift",
		description: "Swift async/await and structured concurrency",
		content: `id: swift-async-await
name: swift-async-await
description: Swift async/await and structured concurrency

content: |
  Swift concurrency best practices:
  - Use async/await for asynchronous code
  - Use Task for launching concurrent work
  - Use TaskGroup for parallel operations
  - Use actors for thread-safe state
  - Mark MainActor for UI updates
  
  Example:
    func fetchData() async throws -> Data {
        let (data, response) = try await URLSession.shared.data(from: url)
        guard let http = response as? HTTPURLResponse,
              http.statusCode == 200 else {
            throw FetchError.badResponse
        }
        return data
    }
    
    // Parallel fetch
    async let users = fetchUsers()
    async let posts = fetchPosts()
    let (u, p) = try await (users, posts)

tags:
  confirmed: [swift, async, concurrency, await]
schema_version: 2
`,
	},

	// General patterns
	{
		name:        "commit-messages",
		category:    "general",
		description: "Conventional commit message format",
		content: `id: commit-messages
name: commit-messages
description: Conventional commit message format

content: |
  Use conventional commits format:
  
  <type>(<scope>): <subject>
  
  Types:
  - feat: New feature
  - fix: Bug fix
  - docs: Documentation only
  - style: Formatting, no code change
  - refactor: Code change, no feature/fix
  - perf: Performance improvement
  - test: Adding tests
  - chore: Build, deps, etc.
  
  Examples:
    feat(auth): add OAuth2 login
    fix(api): handle null response
    docs: update README with examples
    refactor(core): extract validation logic
  
  Rules:
  - Use imperative mood ("add" not "added")
  - No period at end
  - Keep under 72 characters

tags:
  confirmed: [git, commits, best-practices]
schema_version: 2
`,
	},
	{
		name:        "code-review",
		category:    "general",
		description: "Code review checklist and feedback patterns",
		content: `id: code-review
name: code-review
description: Code review checklist and feedback patterns

content: |
  When reviewing code, check:
  
  Functionality:
  - Does it do what it's supposed to?
  - Are edge cases handled?
  - Are errors handled properly?
  
  Quality:
  - Is the code readable and clear?
  - Are names descriptive?
  - Is there unnecessary complexity?
  - Is there code duplication?
  
  Testing:
  - Are there adequate tests?
  - Do tests cover edge cases?
  - Are tests readable?
  
  Feedback style:
  - Be specific and constructive
  - Explain the "why"
  - Suggest alternatives
  - Praise good patterns
  
  Example feedback:
    "Consider using a map here instead of iterating - 
     it would be O(1) lookup instead of O(n)"

tags:
  confirmed: [code-review, best-practices]
schema_version: 2
`,
	},
}

func runExamples(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("📚 Example Patterns")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	categories := make(map[string][]examplePattern)
	for _, ex := range builtinExamples {
		categories[ex.category] = append(categories[ex.category], ex)
	}

	for cat, patterns := range categories {
		fmt.Printf("📁 %s\n", strings.ToUpper(cat))
		for _, p := range patterns {
			fmt.Printf("   • %-25s %s\n", p.name, p.description)
		}
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Install:  mur examples install <category|all>")
	fmt.Println("Preview:  mur examples show <name>")
	fmt.Println()

	return nil
}

func runExamplesInstall(cmd *cobra.Command, args []string) error {
	category := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		return err
	}

	installed := 0
	skipped := 0

	for _, ex := range builtinExamples {
		if category != "all" && ex.category != category {
			continue
		}

		path := filepath.Join(patternsDir, ex.name+".yaml")
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("⏭️  %s (already exists)\n", ex.name)
			skipped++
			continue
		}

		if err := os.WriteFile(path, []byte(ex.content), 0644); err != nil {
			fmt.Printf("❌ %s: %v\n", ex.name, err)
			continue
		}

		fmt.Printf("✅ %s\n", ex.name)
		installed++
	}

	fmt.Println()
	if installed > 0 {
		fmt.Printf("Installed %d patterns\n", installed)
		fmt.Println("Run 'mur sync' to sync to AI tools")
	}
	if skipped > 0 {
		fmt.Printf("Skipped %d (already exist)\n", skipped)
	}

	return nil
}

func runExamplesShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	for _, ex := range builtinExamples {
		if ex.name == name {
			fmt.Println()
			fmt.Printf("📄 %s\n", ex.name)
			fmt.Printf("   Category: %s\n", ex.category)
			fmt.Printf("   %s\n", ex.description)
			fmt.Println()
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println(ex.content)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()
			fmt.Printf("Install: mur examples install %s\n", ex.category)
			return nil
		}
	}

	return fmt.Errorf("example not found: %s", name)
}
