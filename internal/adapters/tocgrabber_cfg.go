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

func DefaultCfg() GrabberCfg {
	return GrabberCfg{
		Path:       "",
		AbsPaths:   false,
		StartDepth: 0,
		Depth:      0,
		Escape:     true,
		Indent:     2,
	}
}
