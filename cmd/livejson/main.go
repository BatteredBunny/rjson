package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/BatteredBunny/rjson"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

func main() {
	var jsonFile string
	flag.StringVar(&jsonFile, "file", "", "Path to JSON file (required)")
	flag.Parse()

	if jsonFile == "" {
		fmt.Fprintf(os.Stderr, "%s JSON file is required\n", red("Error:"))
		flag.Usage()
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(jsonFile))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func loadJSONFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}

type model struct {
	query string

	jsonData []byte
	jsonFile string
}

func (m model) Init() tea.Cmd {
	return nil
}

func initialModel(jsonFile string) model {
	jsonData, err := loadJSONFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("Error loading JSON file:"), err)
		os.Exit(1)
	}

	return model{
		query:    "",
		jsonData: jsonData,
		jsonFile: jsonFile,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		pressed := msg.String()
		switch pressed {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "down", "tab":
			break
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
			}
		case "enter":
			m.query = ""
		default:
			m.query += pressed
		}
	}

	return m, nil
}

func (m model) View() string {
	var s string

	s += fmt.Sprintf("%s - %s | Press q to quit\n", bold("RJSON repl"), green(m.jsonFile))
	s += fmt.Sprintf("%s\n", executeQuery(m.query, m.jsonData))

	s += fmt.Sprintf("%s %s", cyan("query>"), m.query)

	return s
}

func executeQuery(query string, jsonData []byte) string {
	result, err := rjson.QueryJson(jsonData, query)
	if err != nil {
		return fmt.Sprintf("%s %v", red("Query Error:"), err)
	}

	var output []byte
	output, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		output = result
	}

	return green(string(output))
}
