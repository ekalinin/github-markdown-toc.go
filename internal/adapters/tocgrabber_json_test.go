package adapters

import "testing"

func getTestJson() string {
	// how to get example:
	// ❯ curl -s -H 'Content-Type: application/json' -H 'Accept: application/json' \
	// 		https://github.com/ekalinin/sitemap.js/blob/6bc3eb12c898c1037a35a11b2eb24ababdeb3580/README.md | \
	// 		jq .payload.blob.headerInfo.toc
	// [
	// 	{
	// 	  "level": 1,
	// 	  "text": "sitemap.js",
	// 	  "anchor": "sitemapjs",
	// 	  "htmlText": "sitemap.js"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "Installation",
	// 	  "anchor": "installation",
	// 	  "htmlText": "Installation"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "Usage",
	// 	  "anchor": "usage",
	// 	  "htmlText": "Usage"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "License",
	// 	  "anchor": "license",
	// 	  "htmlText": "License"
	// 	}
	//   ]
	return `
	{
		"payload": {
			"blob": {
				"headerInfo": {
					"toc": [
						{
							"level": 1,
							"text": "sitemap.js",
							"anchor": "sitemapjs",
							"htmlText": "sitemap.js"
						},
						{
							"level": 2,
							"text": "Installation",
							"anchor": "installation",
							"htmlText": "Installation"
						},
						{
							"level": 2,
							"text": "Usage",
							"anchor": "usage",
							"htmlText": "Usage"
						},
						{
							"level": 3,
							"text": "Example",
							"anchor": "example",
							"htmlText": "Example"
						},
						{
							"level": 2,
							"text": "License",
							"anchor": "license",
							"htmlText": "License"
						}
					]
				}
			}
		}
	}
	`
}

func Test_JsonGrabberDefaoult(t *testing.T) {
	grabber := NewJsonGrabber(DefaultCfg())
	toc, err := grabber.Grab(getTestJson())
	if err != nil {
		t.Errorf("got error from grabber: %v", err)
	}

	linesWanted := 5
	if len(*toc) != linesWanted {
		t.Errorf("toc is not full (want %d lines, got=%d): %v", linesWanted, len(*toc), *toc)
	}

	tocWanted := []string{
		"* [sitemap\\.js](#sitemapjs)",
		"  * [Installation](#installation)",
		"  * [Usage](#usage)",
		"    * [Example](#example)",
		"  * [License](#license)",
	}

	for i, s := range *toc {
		if s != tocWanted[i] {
			t.Errorf("toc is not correct at i=%d. want=%s, got=%s",
				i, tocWanted[i], s)
		}
	}
}

func Test_JSONGrabberWithOptions(t *testing.T) {
	cfg := DefaultCfg()
	cfg.StartDepth = 1
	cfg.Depth = 2
	cfg.AbsPaths = true
	cfg.Path = "github-markdown-toc.go"
	grabber := NewJsonGrabber(cfg)
	toc, err := grabber.Grab(getTestJson())
	if err != nil {
		t.Errorf("got error from grabber: %v", err)
	}
	linesWanted := 3
	if len(*toc) != linesWanted {
		t.Errorf("toc is not full (want %d lines, got=%d): %v", linesWanted, len(*toc), *toc)
	}
	tocWanted := []string{
		"* [Installation](" + cfg.Path + "#installation)",
		"* [Usage](" + cfg.Path + "#usage)",
		"* [License](" + cfg.Path + "#license)",
	}

	for i, s := range *toc {
		if s != tocWanted[i] {
			t.Errorf("toc is not correct at i=%d. want=%s, got=%s",
				i, tocWanted[i], s)
		}
	}
}

func Test_JSONGrabberFail(t *testing.T) {
	jsonBody := `{`
	grabber := NewJsonGrabber(DefaultCfg())
	_, err := grabber.Grab(jsonBody)
	if err == nil {
		t.Errorf("should fail")
	}
}
