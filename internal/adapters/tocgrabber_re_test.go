package adapters

import (
	"fmt"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/internal/version"
)

const (
	HTML_README_OTHER_LANG_0 = `
	<h1><a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true"><span class="octicon octicon-link"></span></a>README in another language</h1>
	`
	HTML_README_OTHER_LANG_2023_10 = `
	<h1 id="user-content-readme-in-another-language"><a class="heading-link" href="#readme-in-another-language">README in another language<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
	`
	HTML_README_OTHER_LANG_2024_03 = `
	<div class="markdown-heading"><h1 class="heading-element">README in another language</h1><a id="user-content-readme-in-another-language" class="anchor-element" aria-label="Permalink: README in another language" href="#readme-in-another-language"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
	`
)

type reTest struct {
	html    string
	version string
}

func checkTest(t *testing.T, tests []reTest, cfg GrabberCfg, expected []string) {
	for _, d := range tests {
		t.Run(fmt.Sprintf("v.%s", d.version), func(t *testing.T) {
			grabber := NewReGrabber("", cfg, d.version)
			toc, _ := grabber.Grab(d.html)
			if len(*toc) != len(expected) {
				t.Errorf("Rows differs. Got: %d, want: %d (got toc=%v)\n",
					len(*toc), len(expected), *toc)
			}
			for i, got := range *toc {
				want := expected[i]
				if got != want {
					t.Errorf("\nGot     : %s\nExpected: %s\n", got, want)
				}
			}
		})
	}
}

func Test_ReGrabberOneRow(t *testing.T) {
	// https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ grep "README in another" /var/folders/5t/spm0zsl13zx4p0b4z5s01d04qb6th3/T/ghtoc-remote-txt91529502.debug.html
	expected := []string{
		"* [README in another language](#readme-in-another-language)",
	}

	tests := []reTest{
		{
			HTML_README_OTHER_LANG_0,
			version.GH_V0,
		},
		{
			HTML_README_OTHER_LANG_2023_10,
			version.GH_2023_10,
		},
		{
			HTML_README_OTHER_LANG_2024_03,
			version.GH_2024_03,
		},
	}
	checkTest(t, tests, DefaultCfg(), expected)
}

func Test_ReGrabberOneRowWithNewLines(t *testing.T) {
	// https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	expected := []string{
		"* [README in another language](#readme-in-another-language)",
	}
	tests := []reTest{
		{
			`
			<h1 id="user-content-readme-in-another-language">
				<a class="heading-link" href="#readme-in-another-language">
					README in another language
					<span aria-hidden="true" class="octicon octicon-link"></span>
				</a>
			</h1>
			`,
			version.GH_2023_10,
		},
	}
	checkTest(t, tests, DefaultCfg(), expected)
}

func Test_ReGrabberMultilineOriginGithub(t *testing.T) {
	// https://github.com/ekalinin/envirius/blob/master/README.md#how-to-add-a-plugin
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	expected := []string{
		"* [How to add a plugin?](#how-to-add-a-plugin)",
		"  * [Mandatory elements](#mandatory-elements)",
		"    * [plug\\_list\\_versions](#plug_list_versions)",
	}
	tests := []reTest{
		{
			html: `
<h1 id="user-content-how-to-add-a-plugin"><a class="heading-link" href="#how-to-add-a-plugin">How to add a plugin?<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<p>All plugins are in the directory
<a href="https://github.com/ekalinin/envirius/tree/master/src/nv-plugins">nv-plugins</a>.
If you need to add support for a new language you should add it as plugin
inside this directory.</p>
<h2 id="user-content-mandatory-elements"><a class="heading-link" href="#mandatory-elements">Mandatory elements<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<p>If you create a plugin which builds all stuff from source then In a simplest
case you need to implement 2 functions in the plugin's body:</p>
<h3 id="user-content-plug_list_versions"><a class="heading-link" href="#plug_list_versions">plug_list_versions<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
<p>This function should return list of available versions of the plugin.
For example:</p>
	`,
			version: version.GH_2023_10,
		},
		{
			html: `
<div class="markdown-heading"><h1 class="heading-element">How to add a plugin?</h1><a id="user-content-how-to-add-a-plugin" class="anchor-element" aria-label="Permalink: How to add a plugin?" href="#how-to-add-a-plugin"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>All plugins are in the directory
<a href="https://github.com/ekalinin/envirius/tree/master/src/nv-plugins">nv-plugins</a>.
If you need to add support for a new language you should add it as plugin
inside this directory.</p>
<div class="markdown-heading"><h2 class="heading-element">Mandatory elements</h2><a id="user-content-mandatory-elements" class="anchor-element" aria-label="Permalink: Mandatory elements" href="#mandatory-elements"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>If you create a plugin which builds all stuff from source then In a simplest
case you need to implement 2 functions in the plugin's body:</p>
<div class="markdown-heading"><h3 class="heading-element">plug_list_versions</h3><a id="user-content-plug_list_versions" class="anchor-element" aria-label="Permalink: plug_list_versions" href="#plug_list_versions"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>This function should return list of available versions of the plugin.
For example:</p>
	`,
			version: version.GH_2024_03,
		},
	}
	checkTest(t, tests, DefaultCfg(), expected)
}

const (
	HTML_MULTILINE_2023_10 = `
<h1 id="user-content-the-command-foo1"><a class="heading-link" href="#the-command-foo1">The command <code>foo1</code>
<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<p>Blabla...</p>
<h2 id="user-content-the-command-foo2-is-better"><a class="heading-link" href="#the-command-foo2-is-better">The command <code>foo2</code> is better<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<p>Blabla...</p>
<h1 id="user-content-the-command-bar1"><a class="heading-link" href="#the-command-bar1">The command <code>bar1</code>
<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<p>Blabla...</p>
<h2 id="user-content-the-command-bar2-is-better"><a class="heading-link" href="#the-command-bar2-is-better">The command <code>bar2</code> is better<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<p>Blabla...</p>
<h3 id="user-content-the-command-bar3-is-the-best"><a class="heading-link" href="#the-command-bar3-is-the-best">The command <code>bar3</code> is the best<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
<p>Blabla...</p>
		`
	HTML_MULTILINE_2024_03 = `
<div class="markdown-heading"><h1 class="heading-element">The command <code>foo1</code>
</h1><a id="user-content-the-command-foo1" class="anchor-element" aria-label="Permalink: The command foo1" href="#the-command-foo1"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Blabla...</p>
<div class="markdown-heading"><h2 class="heading-element">The command <code>foo2</code> is better</h2><a id="user-content-the-command-foo2-is-better" class="anchor-element" aria-label="Permalink: The command foo2 is better" href="#the-command-foo2-is-better"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Blabla...</p>
<div class="markdown-heading"><h1 class="heading-element">The command <code>bar1</code>
</h1><a id="user-content-the-command-bar1" class="anchor-element" aria-label="Permalink: The command bar1" href="#the-command-bar1"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Blabla...</p>
<div class="markdown-heading"><h2 class="heading-element">The command <code>bar2</code> is better</h2><a id="user-content-the-command-bar2-is-better" class="anchor-element" aria-label="Permalink: The command bar2 is better" href="#the-command-bar2-is-better"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Blabla...</p>
<div class="markdown-heading"><h3 class="heading-element">The command <code>bar3</code> is the best</h3><a id="user-content-the-command-bar3-is-the-best" class="anchor-element" aria-label="Permalink: The command bar3 is the best" href="#the-command-bar3-is-the-best"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Blabla...</p>
		`
)

func Test_ReGrabberBackquoted(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	expected := []string{
		"* [The command foo1](#the-command-foo1)",
		"  * [The command foo2 is better](#the-command-foo2-is-better)",
		"* [The command bar1](#the-command-bar1)",
		"  * [The command bar2 is better](#the-command-bar2-is-better)",
		"    * [The command bar3 is the best](#the-command-bar3-is-the-best)",
	}

	tests := []reTest{
		{
			html:    HTML_MULTILINE_2023_10,
			version: version.GH_2023_10,
		},
		{
			html:    HTML_MULTILINE_2024_03,
			version: version.GH_2024_03,
		},
	}
	checkTest(t, tests, DefaultCfg(), expected)
}

func Test_ReGrabberDepth(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	expected := []string{
		"* [The command foo1](#the-command-foo1)",
		"* [The command bar1](#the-command-bar1)",
	}

	tests := []reTest{
		{
			html:    HTML_MULTILINE_2023_10,
			version: version.GH_2023_10,
		},
		{
			html:    HTML_MULTILINE_2024_03,
			version: version.GH_2024_03,
		},
	}

	cfg := DefaultCfg()
	cfg.Depth = 1
	checkTest(t, tests, cfg, expected)
}

func Test_ReGrabberStartDepth(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	expected := []string{
		"* [The command foo2 is better](#the-command-foo2-is-better)",
		"* [The command bar2 is better](#the-command-bar2-is-better)",
		"  * [The command bar3 is the best](#the-command-bar3-is-the-best)",
	}

	tests := []reTest{
		{
			html:    HTML_MULTILINE_2023_10,
			version: version.GH_2023_10,
		},
		{
			html:    HTML_MULTILINE_2024_03,
			version: version.GH_2024_03,
		},
	}

	cfg := DefaultCfg()
	cfg.StartDepth = 1
	checkTest(t, tests, cfg, expected)
}

func Test_ReGrabberAbsPath(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	link := "https://github.com/ekalinin/envirius/blob/master/README.md"
	expected := []string{
		"* [README in another language](" + link + "#readme-in-another-language)",
	}

	tests := []reTest{
		{
			html:    HTML_README_OTHER_LANG_2023_10,
			version: version.GH_2023_10,
		},
		{
			html:    HTML_README_OTHER_LANG_2024_03,
			version: version.GH_2024_03,
		},
	}
	cfg := DefaultCfg()
	cfg.AbsPaths = true
	cfg.Path = link
	checkTest(t, tests, cfg, expected)
}

func Test_ReGrabberEscapedChars(t *testing.T) {
	expected := []string{
		"* [mod\\_\\*](#mod_)",
	}

	tests := []reTest{
		{
			html: `
			<h2 id="user-content-mod_"><a class="heading-link" href="#mod_">mod_*<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
			`,
			version: version.GH_2023_10,
		},
		{
			html: `
			<div class="markdown-heading"><h2 class="heading-element">mod_*</h2><a id="user-content-mandatory-elements" class="anchor-element" aria-label="Permalink: Mandatory elements" href="#mod_"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
			`,
			version: version.GH_2024_03,
		},
	}
	checkTest(t, tests, DefaultCfg(), expected)
}

func Test_ReGrabberCustomSpaceIndentation(t *testing.T) {
	/*
		$ cat test.md
		# Header Level1
		## Header Level2
		### Header Level3
		$ go run cmd/gh-md-toc/main.go --debug test.md
		$ cat test.md.debug.html
	*/
	expected := []string{
		"* [Header Level1](#header-level1)",
		"    * [Header Level2](#header-level2)",
		"        * [Header Level3](#header-level3)",
	}

	tests := []reTest{
		{
			html: `
<h1 id="user-content-header-level1"><a class="heading-link" href="#header-level1">Header Level1<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<h2 id="user-content-header-level2"><a class="heading-link" href="#header-level2">Header Level2<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<h3 id="user-content-header-level3"><a class="heading-link" href="#header-level3">Header Level3<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
	`,
			version: version.GH_2023_10,
		},
		{
			html: `
<div class="markdown-heading"><h1 class="heading-element">Header Level1</h1><a id="user-content-header-level1" class="anchor-element" aria-label="Permalink: Header Level1" href="#header-level1"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h2 class="heading-element">Header Level2</h2><a id="user-content-header-level2" class="anchor-element" aria-label="Permalink: Header Level2" href="#header-level2"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h3 class="heading-element">Header Level3</h3><a id="user-content-header-level3" class="anchor-element" aria-label="Permalink: Header Level3" href="#header-level3"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
	`,
			version: version.GH_2024_03,
		},
	}
	cfg := DefaultCfg()
	cfg.Indent = 4
	checkTest(t, tests, cfg, expected)
}
