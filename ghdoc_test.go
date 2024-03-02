package ghtoc

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/internal"
)

func TestIsUrl(t *testing.T) {
	doc1 := &GHDoc{
		Path: "https://github.com/ekalinin/envirius/blob/master/README.md",
	}
	if !doc1.IsRemoteFile() {
		t.Error("This is url: ", doc1.Path)
	}

	doc2 := &GHDoc{
		Path: "./README.md",
	}
	if doc2.IsRemoteFile() {
		t.Error("This is not url: ", doc2.Path)
	}
}

func checkTestsOne(tests []*GHDoc, t *testing.T, exptected string) {
	for _, d := range tests {
		t.Run(fmt.Sprintf("v.%s", d.reVersion), func(t *testing.T) {
			toc := *d.GrabToc()
			if toc[0] != exptected {
				t.Error("Res :", toc, "\nExpected     :", exptected)
			}
		})
	}
}

func checkTestsMany(tests []*GHDoc, t *testing.T, exptected []string) {
	for _, d := range tests {
		t.Run(fmt.Sprintf("v.%s", d.reVersion), func(t *testing.T) {
			toc := *d.GrabToc()
			for i := range exptected {
				if toc[i] != exptected[i] {
					t.Error("Res :", toc[i], "\nExpected     :", exptected[i])
				}
			}
		})
	}
}

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

func TestGrabTocOneRow(t *testing.T) {
	// https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ grep "README in another" /var/folders/5t/spm0zsl13zx4p0b4z5s01d04qb6th3/T/ghtoc-remote-txt91529502.debug.html
	tocExpected := []string{
		"* [README in another language](#readme-in-another-language)",
	}
	tests := []*GHDoc{
		{
			html:      HTML_README_OTHER_LANG_0,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_V0,
		},
		{
			html:      HTML_README_OTHER_LANG_2023_10,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html:      HTML_README_OTHER_LANG_2024_03,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}

	checkTestsOne(tests, t, tocExpected[0])
}

func TestGrabTocOneRowWithNewLines(t *testing.T) {
	// https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	tocExpected := []string{
		"* [README in another language](#readme-in-another-language)",
	}
	tests := []*GHDoc{
		{
			html: `
			<h1 id="user-content-readme-in-another-language">
				<a class="heading-link" href="#readme-in-another-language">
					README in another language
					<span aria-hidden="true" class="octicon octicon-link"></span>
				</a>
			</h1>
		`, AbsPaths: false,
			Depth:     0,
			Escape:    true,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
	}

	checkTestsOne(tests, t, tocExpected[0])
}

func TestGrabTocMultilineOriginGithub(t *testing.T) {
	// https://github.com/ekalinin/envirius/blob/master/README.md#how-to-add-a-plugin
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/envirius/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md
	tocExpected := []string{
		"* [How to add a plugin?](#how-to-add-a-plugin)",
		"  * [Mandatory elements](#mandatory-elements)",
		"    * [plug\\_list\\_versions](#plug_list_versions)",
	}
	tests := []*GHDoc{
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
			AbsPaths:  false,
			Escape:    true,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
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
			AbsPaths:  false,
			Escape:    true,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}

	checkTestsMany(tests, t, tocExpected)
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

func TestGrabTocBackquoted(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	tocExpected := []string{
		"* [The command foo1](#the-command-foo1)",
		"  * [The command foo2 is better](#the-command-foo2-is-better)",
		"* [The command bar1](#the-command-bar1)",
		"  * [The command bar2 is better](#the-command-bar2-is-better)",
	}

	tests := []*GHDoc{
		{
			html:      HTML_MULTILINE_2023_10,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html:      HTML_MULTILINE_2024_03,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}

	checkTestsMany(tests, t, tocExpected)
}

func TestGrabTocDepth(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	tocExpected := []string{
		"* [The command foo1](#the-command-foo1)",
		"* [The command bar1](#the-command-bar1)",
	}

	tests := []*GHDoc{
		{
			html:      HTML_MULTILINE_2023_10,
			AbsPaths:  false,
			Escape:    true,
			Depth:     1,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html:      HTML_MULTILINE_2024_03,
			AbsPaths:  false,
			Escape:    true,
			Depth:     1,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}
	checkTestsMany(tests, t, tocExpected)

}

func TestGrabTocStartDepth(t *testing.T) {
	// https://github.com/ekalinin/github-markdown-toc/blob/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	// $ go run cmd/gh-md-toc/main.go --debug https://raw.githubusercontent.com/ekalinin/github-markdown-toc/656b34011a482544a9ebb4116332c044834bdbbf/tests/test%20directory/test_backquote.md
	tocExpected := []string{
		"* [The command foo2 is better](#the-command-foo2-is-better)",
		"* [The command bar2 is better](#the-command-bar2-is-better)",
		"  * [The command bar3 is the best](#the-command-bar3-is-the-best)",
	}

	tests := []*GHDoc{
		{
			html:       HTML_MULTILINE_2023_10,
			AbsPaths:   false,
			Escape:     true,
			StartDepth: 1,
			Indent:     2,
			reVersion:  internal.GH_2023_10,
		},
		{
			html:       HTML_MULTILINE_2024_03,
			AbsPaths:   false,
			Escape:     true,
			StartDepth: 1,
			Indent:     2,
			reVersion:  internal.GH_2024_03,
		},
	}

	checkTestsMany(tests, t, tocExpected)
}

func TestGrabTocWithAbspath(t *testing.T) {
	link := "https://github.com/ekalinin/envirius/blob/master/README.md"
	tocExpected := []string{
		"* [README in another language](" + link + "#readme-in-another-language)",
	}

	tests := []*GHDoc{
		{
			html:      HTML_README_OTHER_LANG_2023_10,
			AbsPaths:  true,
			Path:      link,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html:      HTML_README_OTHER_LANG_2024_03,
			AbsPaths:  true,
			Path:      link,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}

	checkTestsOne(tests, t, tocExpected[0])
}

func TestEscapedChars(t *testing.T) {
	tocExpected := []string{
		"* [mod\\_\\*](#mod_)",
	}

	tests := []*GHDoc{
		{
			html: `
			<h2 id="user-content-mod_"><a class="heading-link" href="#mod_">mod_*<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
			`,
			AbsPaths:  false,
			Escape:    true,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html: `
			<div class="markdown-heading"><h2 class="heading-element">mod_*</h2><a id="user-content-mandatory-elements" class="anchor-element" aria-label="Permalink: Mandatory elements" href="#mod_"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
			`,
			AbsPaths:  false,
			Escape:    true,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}
	checkTestsOne(tests, t, tocExpected[0])
}

func TestCustomSpaceIndentation(t *testing.T) {
	/*
		$ cat test.md
		# Header Level1
		## Header Level2
		### Header Level3

		$ go run cmd/gh-md-toc/main.go --debug test.md
		$ cat test.md.debug.html
	*/
	tocExpected := []string{
		"* [Header Level1](#header-level1)",
		"    * [Header Level2](#header-level2)",
		"        * [Header Level3](#header-level3)",
	}

	tests := []*GHDoc{
		{
			html: `
<h1 id="user-content-header-level1"><a class="heading-link" href="#header-level1">Header Level1<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<h2 id="user-content-header-level2"><a class="heading-link" href="#header-level2">Header Level2<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<h3 id="user-content-header-level3"><a class="heading-link" href="#header-level3">Header Level3<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
	`,
			AbsPaths:  false,
			Depth:     0,
			Indent:    4,
			reVersion: internal.GH_2023_10,
		},
		{
			html: `
<div class="markdown-heading"><h1 class="heading-element">Header Level1</h1><a id="user-content-header-level1" class="anchor-element" aria-label="Permalink: Header Level1" href="#header-level1"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h2 class="heading-element">Header Level2</h2><a id="user-content-header-level2" class="anchor-element" aria-label="Permalink: Header Level2" href="#header-level2"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h3 class="heading-element">Header Level3</h3><a id="user-content-header-level3" class="anchor-element" aria-label="Permalink: Header Level3" href="#header-level3"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
	`,
			AbsPaths:  false,
			Depth:     0,
			Indent:    4,
			reVersion: internal.GH_2024_03,
		},
	}
	checkTestsMany(tests, t, tocExpected)
}

func TestMinHeaderNumber(t *testing.T) {
	tocExpected := []string{
		"* [foo](#foo)",
		"  * [bar](#bar)",
	}

	doc := &GHDoc{
		html: `
		<h3 id="user-content-foo"><a class="heading-link" href="#foo">foo<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
		<h4 id="user-content-bar"><a class="heading-link" href="#bar">bar<span aria-hidden="true" class="octicon octicon-link"></span></a></h4>
		`,
		AbsPaths:  false,
		Depth:     0,
		Indent:    2,
		reVersion: internal.GH_2023_10,
	}
	toc := *doc.GrabToc()

	if toc[0] != tocExpected[0] {
		t.Error("Res :", toc, "\nExpected     :", tocExpected)
	}
}

func TestGHTocPrint(t *testing.T) {
	toc := GHToc{"one", "two"}
	want := "one\ntwo\n\n"
	var got bytes.Buffer
	toc.Print(&got)

	if got.String() != want {
		t.Error("\nGot :", got.String(), "\nWant:", want)
	}
}

func TestNewGHDocWithDebug(t *testing.T) {
	noMatterN := 1
	noMatterS := "test"
	noMatterB := false
	var got bytes.Buffer

	doc := NewGHDoc(noMatterS, noMatterB, noMatterN, noMatterN,
		noMatterB, noMatterS, noMatterN, true)
	doc.logger = log.New(&got, "", 0)

	want := "test"
	doc.d(want)
	if got.String() != want+"\n" {
		t.Error("\nGot :", got.String(), "\nWant:", want)
	}
}

func TestGHDocConvert2HTML(t *testing.T) {
	remotePath := "https://github.com/some/readme.md"
	token := "some-gh-token"
	doc := NewGHDoc(remotePath, true, 0, 0,
		true, token, 4, false)

	// mock for getting remote raw README text
	htmlResponse := []byte("raw md text")
	doc.httpGetter = func(urlPath string) ([]byte, string, error) {
		if urlPath != remotePath {
			t.Error("Wrong urlPath. \nGot :", urlPath, "\nWant:", remotePath)
		}
		return htmlResponse, "text/plain;utf-8", nil
	}

	// mock for converting md to txt
	ghURL := "https://api.github.com/markdown/raw"
	htmlBody := `<h1>header></h1>some text`
	doc.httpPoster = func(urlPath, filePath, token string) (string, error) {
		if urlPath != ghURL {
			if urlPath != remotePath {
				t.Error("Wrong urlPath. \nGot :", urlPath, "\nWant:", ghURL)
			}
		}
		return htmlBody, nil
	}
	if err := doc.Convert2HTML(); err != nil {
		t.Error("Got error:", err)
	}
	if doc.html != htmlBody {
		t.Error("Wrong html. \nGot :", doc.html, "\nWant:", htmlBody)
	}
}

func TestGHDocConvert2HTMLNonPlainText(t *testing.T) {
	remotePath := "https://github.com/some/readme.md"
	token := "some-gh-token"
	doc := NewGHDoc(remotePath, true, 0, 0,
		true, token, 4, false)

	// mock for getting remote raw README text
	htmlResponse := []byte("raw md text")
	doc.httpGetter = func(_ string) ([]byte, string, error) {
		return htmlResponse, "text/html;utf-8", nil
	}
	// should not call converter to HTML
	doc.httpPoster = func(urlPath, filePath, token string) (string, error) {
		t.Error("Should not call httpPost (via convertMd2Html)")
		return "", nil
	}
	if err := doc.Convert2HTML(); err != nil {
		t.Error("Got error:", err)
	}
	if doc.html != string(htmlResponse) {
		t.Error("Wrong html. \nGot :", doc.html, "\nWant:", string(htmlResponse))
	}
}

func TestGHDocConvert2HTMLErrorConvert(t *testing.T) {
	remotePath := "https://github.com/some/readme.md"
	token := "some-gh-token"
	errGet := errors.New("error from http get")
	doc := NewGHDoc(remotePath, true, 0, 0,
		true, token, 4, false)

	// mock for getting remote raw README text
	doc.httpGetter = func(urlPath string) ([]byte, string, error) {
		return nil, "", errGet
	}

	err := doc.Convert2HTML()
	if err == nil {
		t.Error("Should get error from http get!")
	}

	if !errors.Is(err, errGet) {
		t.Error("Wrong error. \nGot :", err, "\nWant:", errGet)
	}
}

func TestGHDocConvert2HTMLLocalFileNotExists(t *testing.T) {
	localPath := "/some/readme.md"
	token := "some-gh-token"
	doc := NewGHDoc(localPath, true, 0, 0,
		true, token, 4, false)

	// should not be called
	doc.httpGetter = func(_ string) ([]byte, string, error) {
		t.Error("Should not call httpGet")
		return nil, "", nil
	}

	err := doc.Convert2HTML()
	if err == nil {
		t.Error("Should get error from file checking.")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Error("Wrong error. \nGot :", err, "\nWant:", os.ErrNotExist)
	}
}

// Cover the changes of `ioutil.*` to `os.*` in Convert2HTML.
func TestGHDocConvert2HTML_issue35(t *testing.T) {
	remotePath := "https://github.com/some/readme.md"
	token := "some-gh-token"

	// enable debug
	doc := NewGHDoc(remotePath, true, 0, 0, true, token, 4, false)

	// mock for getting remote raw README text
	htmlResponse := []byte("raw md text")
	doc.httpGetter = func(urlPath string) ([]byte, string, error) {
		return htmlResponse, "text/plain;utf-8", nil
	}

	// mock for converting md to txt
	htmlBody := `<h1>header></h1>some text`
	doc.httpPoster = func(urlPath, filePath, token string) (string, error) {
		return htmlBody, nil
	}

	if err := doc.Convert2HTML(); err != nil {
		t.Error("Got error:", err)
	}

	if doc.html != htmlBody {
		t.Error("Wrong html. \nGot :", doc.html, "\nWant:", htmlBody)
	}
}

func TestGrabToc_issue35(t *testing.T) {
	/*
		$ cat test.md
		# One
		## Two
		### Three

		$ go run cmd/gh-md-toc/main.go --debug test.md
		$ cat test.md.debug.html
	*/
	tocExpected := []string{
		"* [One](#one)",
		"  * [Two](#two)",
		"    * [Three](#three)",
	}

	tests := []*GHDoc{
		{
			html: `
<h1 id="user-content-one"><a class="heading-link" href="#one">One<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<h2 id="user-content-two"><a class="heading-link" href="#two">Two<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<h3 id="user-content-three"><a class="heading-link" href="#three">Three<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>
`,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2023_10,
		},
		{
			html: `
<div class="markdown-heading"><h1 class="heading-element">One</h1><a id="user-content-one" class="anchor-element" aria-label="Permalink: One" href="#one"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h2 class="heading-element">Two</h2><a id="user-content-two" class="anchor-element" aria-label="Permalink: Two" href="#two"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h3 class="heading-element">Three</h3><a id="user-content-three" class="anchor-element" aria-label="Permalink: Three" href="#three"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
`,
			AbsPaths:  false,
			Depth:     0,
			Indent:    2,
			reVersion: internal.GH_2024_03,
		},
	}

	checkTestsMany(tests, t, tocExpected)
}

func TestSetGHURL(t *testing.T) {
	noSense := "xxx"
	doc := NewGHDoc(noSense, true, 0, 0, true, noSense, 4, true)

	ghURL := "https://api.github.com"
	if doc.ghURL != ghURL {
		t.Error("Res :", doc.ghURL, "\nExpected     :", ghURL)
	}

	ghURL = "https://api.xxx.com"
	doc.SetGHURL(ghURL)
	if doc.ghURL != ghURL {
		t.Error("Res :", doc.ghURL, "\nExpected     :", ghURL)
	}

	// mock for converting md to txt (just to check passing new GH URL)
	doc.httpPoster = func(urlPath, filePath, token string) (string, error) {
		ghURLFull := ghURL + "/markdown/raw"
		if urlPath != ghURLFull {
			t.Error("Res :", urlPath, "\nExpected     :", ghURL)
		}
		return noSense, nil
	}

	if _, err := doc.convertMd2Html(noSense, noSense); err != nil {
		t.Error("Convert error:", err)
	}
}
