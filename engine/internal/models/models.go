package models

type LogEvent struct {
	Type    string `json:"type"`
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type Config struct {
	TargetDir      string `json:"targetDir"`
	UseAI          bool   `json:"useAI"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"baseUrl"`
	APIKey         string `json:"apiKey"`
	Model          string `json:"model"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	FocusArea      string `json:"focusArea"`
	StruggleArea   string `json:"struggleArea"`
	HumanErrors    bool   `json:"humanErrors"`
	ASTPhasing     bool   `json:"astPhasing"`
	DepAlignment   bool   `json:"depAlignment"`
	Branches       bool   `json:"branches"`
}

type Estimate struct {
	Commits  int    `json:"commits"`
	Branches int    `json:"branches"`
	Runtime  string `json:"runtime"`
	Size     string `json:"size"`
}

type State struct {
	Status     string `json:"status"`
	Before     int    `json:"before"`
	After      int    `json:"after"`
	Verified   bool   `json:"verified"`
	ReportPath string `json:"report_path"`
}
