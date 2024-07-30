# gq

`gq` is a command-line tool for querying HTML files using [goquery](https://github.com/PuerkitoBio/goquery), a Go implementation of jQuery. This tool allows you to extract specific information from HTML documents using CSS selectors combined with custom chained commands. The primary purpose of gq is to serve as an experimental platform for exploring and showcasing the capabilities of [goquery](https://github.com/PuerkitoBio/goquery) in practical applications.

## Installation

To install `gq`, make sure you have Go installed on your system, then run:

```
go install github.com/viswesr/gq@latest
```

## Usage

```
gq -file=<html_file> -query=<query>
```

- `<html_file>`: Path to the HTML file you want to query
- `<query>`: The query string to execute on the HTML file

## Query Language

The query language consists of chained commands separated by the `|` character. The Each block uses `{` and `}`. The basic structure is:

```
Find <selector>|<command1>|<command2>|...
```

### Available Commands

- `Find <selector>`: Select elements using CSS selectors
- `Each{<subquery>}`: Execute a subquery for each selected element
- `First`: Select only the first element of the current selection
- `Last`: Select only the last element of the current selection
- `Parent`: Select the parent element of the current selection
- `Children`: Select the children elements of the current selection
- `Attrib <attribute>`: Get the value of the specified attribute
- `Text`: Get the text content of the selected element(s)
- `Html`: Get the inner HTML of the selected element(s)
- `OuterHtml`: Get the outer HTML of the selected element(s)

## Examples

### Simple Queries

1. Get all links on a page:
   ```
   Find a|Each{Attrib href}
   ```

2. Get the text of all paragraphs:
   ```
   Find p|Each{Text}
   ```

3. Get the title of the page:
   ```
   Find title|Text
   ```

4. Get the value of all input fields:
   ```
   Find input|Each{Attrib value}
   ```

5. Get the src of all images:
   ```
   Find img|Each{Attrib src}
   ```

### Complex Queries

1. Get the href of the first link within each div with class "content":
   ```
   Find div.content|Each{Find a|First|Attrib href}
   ```

2. Get the text of the last list item in an unordered list with id "menu":
   ```
   Find ul#menu|Find li|Last|Text
   ```

3. Get the HTML content of all divs with class "article", but only the text of their first paragraph:
   ```
   Find div.article|Each{OuterHtml|Find p|First|Text}
   ```

4. Get the value of the "data-id" attribute for all buttons inside a form with class "login":
   ```
   Find form.login|Find button|Each{Attrib data-id}
   ```

5. Get the href of all links within the nav element, along with their text:
   ```
   Find nav a|Each{Attrib href|Text}
   ```

6. Get the source of all iframes, along with their width and height:
   ```
   Find iframe|Each{Attrib src|Attrib width|Attrib height}
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.
