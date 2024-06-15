package adapters

type GrabberCfg struct {
	Path string

	// toc grabber
	AbsPaths   bool
	StartDepth int
	Depth      int
	Escape     bool
	Indent     int
}
