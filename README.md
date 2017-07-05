github-markdown-toc
===================

This is a golang based implementation of the
[github-markdown-toc](https://github.com/ekalinin/github-markdown-toc) tool.

The advantages of this implementation:

  * no dependencies (no need curl, wget, awk, etc.)
  * cross-platform (support for Windows, Mac OS, etc.)
  * regexp for parsing TOC
  * parallel processing of multiple documents


*Attention*: gh-md-toc is able to work properly only if your machine is
connected to the Internet.

Table of Contents
=================

  * [github-markdown-toc](#github-markdown-toc)
  * [Installation](#installation)
    * [Precompiled binaries](#precompiled-binaries)
    * [Compiling from source](#compiling-from-source)
  * [Tests](#tests)
  * [Usage](#usage)
    * [STDIN](#stdin)
    * [Local files](#local-files)
    * [Remote files](#remote-files)
    * [Multiple files](#multiple-files)
    * [Combo](#combo)
    * [Depth](#depth)
  * [LICENSE](#license)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

Installation
============

Precompiled binaries
--------------------

See the releases page, "Downloads" section:

  * https://github.com/ekalinin/github-markdown-toc.go/releases

For example:

```bash
$ wget https://github.com/ekalinin/github-markdown-toc.go/releases/download/0.4.0/gh-md-toc.linux.amd64.tgz
$ tar xzvf gh-md-toc.linux.amd64.tgz
gh-md-toc
$ ./gh-md-toc --version
0.4.0
```

Compiling from source
---------------------

You need golang installed is your OS:

```bash
$ make get-deps
$ make build
$ ./gh-md-toc --help
usage: gh-md-toc [<flags>] [<path>...]

Flags:
  --help     Show help (also see --help-long and --help-man).
  --version  Show application version.
  --depth    How many levels of headings to include. Defaults to 0 (all)

Args:
  [<path>]  Local path or URL of the document to grab TOC

```

Tests
=====

```bash
$ make test
coverage: 28.8% of statements
ok      _~/projects/my/github-toc.go    0.003s
```

Usage
=====

STDIN
-----

Here's an example of TOC creating for markdown from STDIN:

```bash
➥ cat ~/projects/Dockerfile.vim/README.md | ./gh-md-toc
  * [Dockerfile.vim](#dockerfilevim)
  * [Screenshot](#screenshot)
  * [Installation](#installation)
        * [OR using Pathogen:](#or-using-pathogen)
        * [OR using Vundle:](#or-using-vundle)
  * [License](#license)
```

Local files
-----------

Here's an example of TOC creating for a local README.md:

```bash
➥ ./gh-md-toc ~/projects/Dockerfile.vim/README.md                                                                                                                                                Вс. марта 22 22:51:46 MSK 2015

Table of Contents
=================

  * [Dockerfile.vim](#dockerfilevim)
  * [Screenshot](#screenshot)
  * [Installation](#installation)
        * [OR using Pathogen:](#or-using-pathogen)
        * [OR using Vundle:](#or-using-vundle)
  * [License](#license)
```

Remote files
------------

And here's an example, when you have a README.md like this:

  * [README.md without TOC](https://github.com/ekalinin/envirius/blob/f939d3b6882bfb6ecb28ef7b6e62862f934ba945/README.md)

And you want to generate TOC for it.

There is nothing easier:

```bash
➥ ./gh-md-toc https://github.com/ekalinin/envirius/blob/master/README.md

Table of Contents
=================

  * [envirius](#envirius)
    * [Idea](#idea)
    * [Features](#features)
  * [Installation](#installation)
  * [Uninstallation](#uninstallation)
  * [Available plugins](#available-plugins)
  * [Usage](#usage)
    * [Check available plugins](#check-available-plugins)
    * [Check available versions for each plugin](#check-available-versions-for-each-plugin)
    * [Create an environment](#create-an-environment)
    * [Activate/deactivate environment](#activatedeactivate-environment)
      * [Activating in a new shell](#activating-in-a-new-shell)
      * [Activating in the same shell](#activating-in-the-same-shell)
    * [Get list of environments](#get-list-of-environments)
    * [Get current activated environment](#get-current-activated-environment)
    * [Do something in environment without enabling it](#do-something-in-environment-without-enabling-it)
    * [Get help](#get-help)
    * [Get help for a command](#get-help-for-a-command)
  * [How to add a plugin?](#how-to-add-a-plugin)
    * [Mandatory elements](#mandatory-elements)
      * [plug_list_versions](#plug_list_versions)
      * [plug_url_for_download](#plug_url_for_download)
      * [plug_build](#plug_build)
    * [Optional elements](#optional-elements)
      * [Variables](#variables)
      * [Functions](#functions)
    * [Examples](#examples)
  * [Example of the usage](#example-of-the-usage)
  * [Dependencies](#dependencies)
  * [Supported OS](#supported-os)
  * [Tests](#tests)
  * [Version History](#version-history)
  * [License](#license)
  * [README in another language](#readme-in-another-language)
```

That's all! Now all you need — is copy/paste result from console into original
README.md.

And here is a result:

  * [README.md with TOC](https://github.com/ekalinin/envirius/blob/24ea3be0d3cc03f4235fa4879bb33dc122d0ae29/README.md)


Multiple files
--------------

It supports multiple files as well:

```bash
➥ ./gh-md-toc \
    https://github.com/aminb/rust-for-c/blob/master/hello_world/README.md \
    https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md \
    https://github.com/aminb/rust-for-c/blob/master/primitive_types_and_operators/README.md \
    https://github.com/aminb/rust-for-c/blob/master/unique_pointers/README.md

  * [Hello world](https://github.com/aminb/rust-for-c/blob/master/hello_world/README.md#hello-world)

  * [Control Flow](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#control-flow)
    * [If](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#if)
    * [Loops](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#loops)
    * [For loops](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#for-loops)
    * [Switch/Match](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#switchmatch)
    * [Method call](https://github.com/aminb/rust-for-c/blob/master/control_flow/README.md#method-call)

  * [Primitive Types and Operators](https://github.com/aminb/rust-for-c/blob/master/primitive_types_and_operators/README.md#primitive-types-and-operators)

  * [Unique Pointers](https://github.com/aminb/rust-for-c/blob/master/unique_pointers/README.md#unique-pointers)
```

Processing of multiple documents is in parallel mode since version 0.4.0
You can use (old) serial mode by passing option ``--serial`` in the console:

```bash
$ ./gh-md-toc --serial ...
```

Timings:

```bash
➥ time (./gh-md-toc --serial README.md ../envirius/README.ru.md ../github-toc/README.md > /dev/null)

real    0m1.200s
user    0m0.040s
sys     0m0.004s
```

```bash
➥ time (./gh-md-toc README.md ../envirius/README.ru.md ../github-toc/README.md > /dev/null)

real    0m0.784s
user    0m0.036s
sys     0m0.004s
```


Combo
-----

You can easily combine both ways:

```bash
➥ ./gh-md-toc \
    ~/projects/Dockerfile.vim/README.md \
    https://github.com/ekalinin/sitemap.s/blob/master/README.md

  * [Dockerfile.vim](~/projects/Dockerfile.vim/README.md#dockerfilevim)
  * [Screenshot](~/projects/Dockerfile.vim/README.md#screenshot)
  * [Installation](~/projects/Dockerfile.vim/README.md#installation)
        * [OR using Pathogen:](~/projects/Dockerfile.vim/README.md#or-using-pathogen)
        * [OR using Vundle:](~/projects/Dockerfile.vim/README.md#or-using-vundle)
  * [License](~/projects/Dockerfile.vim/README.md#license)

  * [sitemap.js](https://github.com/ekalinin/sitemap.js/blob/master/README.md#sitemapjs)
    * [Installation](https://github.com/ekalinin/sitemap.js/blob/master/README.md#installation)
    * [Usage](https://github.com/ekalinin/sitemap.js/blob/master/README.md#usage)
    * [License](https://github.com/ekalinin/sitemap.js/blob/master/README.md#license)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)
```

Depth
-----

Use `--depth=INT` to control how many levels of headers to include in the TOC

```bash
➥ ./gh-md-toc --depth=1 ~/projects/Dockerfile.vim/README.md

Table of Contents
=================

  * [Dockerfile\.vim](#dockerfilevim)
  * [Screenshot](#screenshot)
  * [Installation](#installation)
  * [License](#license)
```

No escape
---------

```
➥ ./gh-md-toc ~/projects/my/Dockerfile.vim/README.md | grep Docker
* [Dockerfile\.vim](#dockerfilevim)

➥ ./gh-md-toc --no-escape ~/projects/my/Dockerfile.vim/README.md | grep Docker
* [Dockerfile.vim](#dockerfilevim)
```

LICENSE
=======

See [LICENSE](https://github.com/ekalinin/github-markdown-toc.go/blob/master/LICENSE)
file.
