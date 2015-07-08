package main

import (
	"testing"
)

func Test_is_url(t *testing.T) {
	url1 := "https://github.com/ekalinin/envirius/blob/master/README.md"
	if !IsUrl(url1) {
		t.Error("This is url: ", url1)
	}

	url2 := "./README.md"
	if IsUrl(url2) {
		t.Error("This is not url: ", url2)
	}
}

func Test_grab_toc_onerow(t *testing.T) {
	t.Skip("not ready yet")
	toc_expected := []string{
		"  * [README in another language](#readme-in-another-language)",
	}
	toc := GrabToc(`
	<h1><a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true"><span class="octicon octicon-link"></span></a>README in another language</h1>
	`)
	if toc[0] != toc_expected[0] {
		t.Error("Res :", toc, "\nExpected     :", toc_expected)
	}
}

func Test_grab_toc_onerow_with_newlines(t *testing.T) {
	t.Skip("not ready yet")
	toc_expected := []string{
		"  * [README in another language](#readme-in-another-language)",
	}
	toc := GrabToc(`
	<h1>
		<a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true">
			<span class="octicon octicon-link"></span>
		</a>
		README in another language
	</h1>
	`)
	if toc[0] != toc_expected[0] {
		t.Error("Res :", toc, "\nExpected     :", toc_expected)
	}
}

func Test_grab_toc_multiline_origin_github(t *testing.T) {

	toc_expected := []string{
		"  * [How to add a plugin?](#how-to-add-a-plugin)",
		"    * [Mandatory elements](#mandatory-elements)",
		"      * [plug_list_versions](#plug_list_versions)",
	}
	toc := GrabToc(`
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
	`)
	for i := 1; i <= len(toc_expected)-1; i++ {
		if toc[i] != toc_expected[i] {
			t.Error("Res :", toc[i], "\nExpected     :", toc_expected[i])
		}
	}
}
