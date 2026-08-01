package entity

// Heading represents a parsed document heading before TOC rendering.
type Heading struct {
	Level  int
	Text   string
	Anchor string
}
