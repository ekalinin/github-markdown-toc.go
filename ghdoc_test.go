package ghtoc

import (
	"bytes"
	"errors"
	"log"
	"os"
	"testing"
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

func TestGrabTocOneRow(t *testing.T) {
	tocExpected := []string{
		"* [README in another language](#readme-in-another-language)",
	}
	doc := &GHDoc{
		html: `
		<h1><a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true"><span class="octicon octicon-link"></span></a>README in another language</h1>
		`,
		AbsPaths: false,
		Depth:    0,
		Indent:   2,
	}
	toc := *doc.GrabToc()
	if toc[0] != tocExpected[0] {
		t.Error("Res :", toc, "\nExpected     :", tocExpected)
	}
}

func TestGrabTocOneRowWithNewLines(t *testing.T) {
	tocExpected := []string{
		"* [README in another language](#readme-in-another-language)",
	}
	doc := &GHDoc{
		html: `
	<h1>
		<a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true">
			<span class="octicon octicon-link"></span>
		</a>
		README in another language
	</h1>
	`, AbsPaths: false,
		Depth:  0,
		Escape: true,
		Indent: 2,
	}
	toc := *doc.GrabToc()
	if toc[0] != tocExpected[0] {
		t.Error("Res :", toc, "\nExpected     :", tocExpected)
	}
}

func TestGrabTocMultilineOriginGithub(t *testing.T) {

	tocExpected := []string{
		"* [How to add a plugin?](#how-to-add-a-plugin)",
		"  * [Mandatory elements](#mandatory-elements)",
		"    * [plug\\_list\\_versions](#plug_list_versions)",
	}
	doc := &GHDoc{
		html: `
<h1><a id="user-content-how-to-add-a-plugin" class="anchor" href="#how-to-add-a-plugin" aria-hidden="true"><span class="octicon octicon-link"></span></a>How to add a plugin?</h1>

<p>All plugins are in the directory
<a href="https://github.com/ekalinin/envirius/tree/master/src/nv-plugins">nv-plugins</a>.
If you need to add support for a new language you should add it as plugin
inside this directory.</p>

<h2><a id="user-content-mandatory-elements" class="anchor" href="#mandatory-elements" aria-hidden="true"><span class="octicon octicon-link"></span></a>Mandatory elements</h2>

<p>If you create a plugin which builds all stuff from source then In a simplest
case you need to implement 2 functions in the plugin's body:</p>

<h3><a id="user-content-plug_list_versions" class="anchor" href="#plug_list_versions" aria-hidden="true"><span class="octicon octicon-link"></span></a>plug_list_versions</h3>

<p>This function should return list of available versions of the plugin.
For example:</p>
	`, AbsPaths: false,
		Escape: true,
		Depth:  0,
		Indent: 2,
	}
	toc := *doc.GrabToc()
	for i := 0; i <= len(tocExpected)-1; i++ {
		if toc[i] != tocExpected[i] {
			t.Error("Res :", toc[i], "\nExpected     :", tocExpected[i])
		}
	}
}

func TestGrabTocBackquoted(t *testing.T) {
	tocExpected := []string{
		"* [The command foo1](#the-command-foo1)",
		"  * [The command foo2 is better](#the-command-foo2-is-better)",
		"* [The command bar1](#the-command-bar1)",
		"  * [The command bar2 is better](#the-command-bar2-is-better)",
	}

	doc := &GHDoc{
		html: `
<h1>
<a id="user-content-the-command-foo1" class="anchor" href="#the-command-foo1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-foo2-is-better" class="anchor" href="#the-command-foo2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo2</code> is better</h2>

<p>Blabla...</p>

<h1>
<a id="user-content-the-command-bar1" class="anchor" href="#the-command-bar1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-bar2-is-better" class="anchor" href="#the-command-bar2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar2</code> is better</h2>

<p>Blabla...</p>
	`, AbsPaths: false,
		Depth:  0,
		Indent: 2,
	}
	toc := *doc.GrabToc()
	for i := 0; i <= len(tocExpected)-1; i++ {
		if toc[i] != tocExpected[i] {
			t.Error("Res :", toc[i], "\nExpected      :", tocExpected[i])
		}
	}
}

func TestGrabTocDepth(t *testing.T) {
	tocExpected := []string{
		"* [The command foo1](#the-command-foo1)",
		"* [The command bar1](#the-command-bar1)",
	}

	doc := &GHDoc{
		html: `
<h1>
<a id="user-content-the-command-foo1" class="anchor" href="#the-command-foo1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-foo2-is-better" class="anchor" href="#the-command-foo2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo2</code> is better</h2>

<p>Blabla...</p>

<h1>
<a id="user-content-the-command-bar1" class="anchor" href="#the-command-bar1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-bar2-is-better" class="anchor" href="#the-command-bar2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar2</code> is better</h2>

<p>Blabla...</p>
	`, AbsPaths: false,
		Escape: true,
		Depth:  1,
		Indent: 2,
	}
	toc := *doc.GrabToc()

	for i := 0; i <= len(tocExpected)-1; i++ {
		if toc[i] != tocExpected[i] {
			t.Error("Res :", toc[i], "\nExpected      :", tocExpected[i])
		}
	}
}

func TestGrabTocStartDepth(t *testing.T) {
	tocExpected := []string{
		"* [The command foo2 is better](#the-command-foo2-is-better)",
		"  * [The command foo3 is even betterer](#the-command-foo3-is-even-betterer)",
		"* [The command bar2 is better](#the-command-bar2-is-better)",
		"  * [The command bar3 is even betterer](#the-command-bar3-is-even-betterer)",
	}

	doc := &GHDoc{
		html: `
<h1>
<a id="user-content-the-command-foo1" class="anchor" href="#the-command-foo1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-foo2-is-better" class="anchor" href="#the-command-foo2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo2</code> is better</h2>

<p>Blabla...</p>

<h3>
<a id="user-content-the-command-foo3-is-even-betterer" class="anchor" href="#the-command-foo3-is-even-betterer" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>foo3</code> is even betterer</h2>

<p>Blabla...</p>

<h1>
<a id="user-content-the-command-bar1" class="anchor" href="#the-command-bar1" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar1</code>
</h1>

<p>Blabla...</p>

<h2>
<a id="user-content-the-command-bar2-is-better" class="anchor" href="#the-command-bar2-is-better" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar2</code> is better</h2>

<p>Blabla...</p>

<h3>
<a id="user-content-the-command-bar3-is-even-betterer" class="anchor" href="#the-command-bar3-is-even-betterer" aria-hidden="true"><span class="octicon octicon-link"></span></a>The command <code>bar3</code> is even betterer</h2>

<p>Blabla...</p>
	`, AbsPaths: false,
		Escape:     true,
		StartDepth: 1,
		Indent:     2,
	}
	toc := *doc.GrabToc()

	for i := 0; i <= len(tocExpected)-1; i++ {
		if toc[i] != tocExpected[i] {
			t.Error("Res :", toc[i], "\nExpected      :", tocExpected[i])
		}
	}
}

func TestGrabTocWithAbspath(t *testing.T) {
	link := "https://github.com/ekalinin/envirius/blob/master/README.md"
	tocExpected := []string{
		"* [README in another language](" + link + "#readme-in-another-language)",
	}
	doc := &GHDoc{
		html: `
	<h1><a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true"><span class="octicon octicon-link"></span></a>README in another language</h1>
	`, AbsPaths: true,
		Path:   link,
		Depth:  0,
		Indent: 2,
	}
	toc := *doc.GrabToc()
	if toc[0] != tocExpected[0] {
		t.Error("Res :", toc, "\nExpected     :", tocExpected)
	}
}

func TestEscapedChars(t *testing.T) {
	tocExpected := []string{
		"* [mod\\_\\*](#mod_)",
	}

	doc := &GHDoc{
		html: `
		<h2>
			<a id="user-content-mod_" class="anchor"
			    href="#mod_" aria-hidden="true">
				<span class="octicon octicon-link"></span>
			</a>
			mod_*
		</h2>`,
		AbsPaths: false,
		Escape:   true,
		Depth:    0,
		Indent:   2,
	}
	toc := *doc.GrabToc()

	if toc[0] != tocExpected[0] {
		t.Error("Res :", toc, "\nExpected     :", tocExpected)
	}
}

func TestCustomSpaceIndentation(t *testing.T) {
	tocExpected := []string{
		"* [Header Level1](#header-level1)",
		"    * [Header Level2](#header-level2)",
		"        * [Header Level3](#header-level3)",
	}

	doc := &GHDoc{
		html: `
<h1>
<a id="user-content-the-command-level1" class="anchor" href="#header-level1" aria-hidden="true"><span class="octicon octicon-link"></span></a>Header Level1
</h1>
<h2>
<a id="user-content-the-command-level2" class="anchor" href="#header-level2" aria-hidden="true"><span class="octicon octicon-link"></span></a>Header Level2
</h2>
<h3>
<a id="user-content-the-command-level3" class="anchor" href="#header-level3" aria-hidden="true"><span class="octicon octicon-link"></span></a>Header Level3
</h3>
	`,
		AbsPaths: false,
		Depth:    0,
		Indent:   4,
	}
	toc := *doc.GrabToc()

	for i := 0; i <= len(tocExpected)-1; i++ {
		if toc[i] != tocExpected[i] {
			t.Error("Res :", toc[i], "\nExpected      :", tocExpected[i])
		}
	}
}

func TestMinHeaderNumber(t *testing.T) {
	tocExpected := []string{
		"* [foo](#foo)",
		"  * [bar](#bar)",
	}

	doc := &GHDoc{
		html: `
		<h3>
			<a id="user-content-" class="anchor" href="#foo" aria-hidden="true">
				<span class="octicon octicon-link"></span>
			</a>
			foo
		</h3>
		<h4>
			<a id="user-content-" class="anchor" href="#bar" aria-hidden="true">
				<span class="octicon octicon-link"></span>
			</a>
			bar
		</h3>
		`,
		AbsPaths: false,
		Depth:    0,
		Indent:   2,
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
	doc := NewGHDoc(remotePath, true, 0, 0, true, token, 4, true)

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
	// As of 2022-08-25, GitHub API returns the HTML in the below format.
	doc := &GHDoc{
		html: `
<h1><a id="user-content-one" class="anchor" aria-hidden="true" href="#one"><span aria-hidden="true" class="octicon octicon-link"></span></a>One</h1>
<p>Uno</p>
<h2><a id="user-content-two" class="anchor" aria-hidden="true" href="#two"><span aria-hidden="true" class="octicon octicon-link"></span></a>Two</h2>
<p>Dos</p>
<h3><a id="user-content-three" class="anchor" aria-hidden="true" href="#three"><span aria-hidden="true" class="octicon octicon-link"></span></a>Three</h3>
<p>Tres</p>`,
		AbsPaths: false,
		Depth:    0,
		Indent:   2,
	}

	tocExpected := []string{
		"* [One](#one)",
		"  * [Two](#two)",
		"    * [Three](#three)",
	}
	toc := *doc.GrabToc()

	// Require not empty
	if len(toc) == 0 {
		t.Fatal("returned ToC is empty. GrabToc could not parse the HTML")
	}

	// Assert equal
	for i, tocActual := range toc {
		if tocExpected[i] != tocActual {
			t.Error("Res :", tocActual, "\nExpected     :", tocExpected)
		}
	}
}
