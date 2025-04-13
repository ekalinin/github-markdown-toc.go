package config

type Config struct {
	Serial       bool
	HideHeader   bool
	HideFooter   bool
	StartDepth   int
	Depth        int
	NoEscape     bool
	Indent       int
	Debug        bool
	GHToken      string
	GHUrl        string
	GHVersion    string
	AbsPathInToc bool
}
