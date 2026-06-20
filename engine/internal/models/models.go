package models

type LogEvent struct {
	Type    string      `json:"type"`
	Level   string      `json:"level,omitempty"`
	Message string      `json:"message,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type Config struct {
	SourceDir      string `json:"sourceDir"`
	OutputDir      string `json:"outputDir"`
	UseAI          bool   `json:"useAI"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"baseUrl"`
	APIKey         string `json:"apiKey"`
	Model          string `json:"model"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	Cadence        string `json:"cadence"` // Low/Medium/High
	FocusArea      string `json:"focusArea"`
	StruggleArea   string `json:"struggleArea"`
	HumanErrors    bool   `json:"humanErrors"`
	ASTPhasing     bool   `json:"astPhasing"`
	DepAlignment   bool   `json:"depAlignment"`
	Branches       bool   `json:"branches"`
}

type ScanResult struct {
	IsGit        bool    `json:"isGit"`
	FileCount    int     `json:"fileCount"`
	FolderCount  int     `json:"folderCount"`
	SizeMB       float64 `json:"sizeMb"`
	CommitCount  int     `json:"commitCount"`
	FirstCommit  string  `json:"firstCommit"`
	LatestCommit string  `json:"latestCommit"`
	BranchCount  int     `json:"branchCount"`
}

type Estimate struct {
	Commits      int    `json:"commits"`
	Branches     int    `json:"branches"`
	PullRequests int    `json:"pullRequests"`
	Versions     int    `json:"versions"`
	Runtime      string `json:"runtime"`
	SizeIncrease string `json:"sizeIncrease"`
}

type State struct {
	Status     string `json:"status"`
	Before     int    `json:"before"`
	After      int    `json:"after"`
	Verified   bool   `json:"verified"`
	OutputPath string `json:"output_path"`
	ReportPath string `json:"report_path"`
}
