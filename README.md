# gq

`gq` is a command-line tool that uses [goquery](https://github.com/PuerkitoBio/goquery) (A Go's jQuery implementation) to query and extract data from HTML files using CSS selectors and chained commands. It serves as an experimental platform for goquery capabilities.

## Installation

To install `gq`, make sure you have Go installed on your system, then run:

```
go install github.com/viswesr/gq@latest
```

## Usage

```
gq -file=<html_file> -query=<query>
```

- `<html_file>`: Path to the HTML file you want to query or web URL
- `<query>`: The query string to execute on the HTML file

### Optional

There is an optional `-gencode` flag that generate go code for the commandline query instead of fetching information from HTML:

```
gq -file=<html_file> -query=<query> -gencode
```


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

5. Get the href of all links within the nav element and print their text:
   ```
   Find nav a|Each{Attrib href|Text}
   ```

6. Get the source of all iframes:
   ```
   Find iframe|Each{Attrib src}
   ```
### Queries using sample.html

1. Extract the main title:
```
gq -file=sample.html -query="Find #main-title|Text"
```

2. List all navigation links:
```
gq -file=sample.html -query="Find nav ul li a|Each{Attrib href}"
```

3. Extract all book titles:
```
gq -file=sample.html -query="Find .book h3|Each{Text}"
```

4. Get the title and author of each book (using two separate queries):
For titles:
```
gq -file=sample.html -query="Find .book h3|Each{Text}"
```
For authors:
```
gq -file=sample.html -query="Find .book .author|Each{Text}"
```

5. List all genres for all books (using two levels of Each):
```
gq -file=sample.html -query="Find .book|Each{Find .genres li|Each{Text}}"
```

6. Get the HTML content of each book in the Fiction section:
```
gq -file=sample.html -query="Find #fiction .book|Each{OuterHtml}"
```

7. Extract the HTML content of the Non-Fiction section:
```
gq -file=sample.html -query="Find #non-fiction|OuterHtml"
```

8. Get the titles and prices of books (using two separate queries):
For titles:
```
gq -file=sample.html -query="Find .book h3|Each{Text}"
```
For prices:
```
gq -file=sample.html -query="Find .book .price|Each{Text}"
```

9. List all section IDs:
```
gq -file=sample.html -query="Find main section|Each{Attrib id}"
```

10. Extract the footer content:
```
gq -file=sample.html -query="Find footer|Text"
```

11. Get the headings of each section and count of books:
```
gq -file=sample.html -query="Find main section|Each{Find h2|Text}"
```
And separately:

```
gq -file=sample.html -query="Find main section|Each{Find .book-list|Children}"
```

12. Extract all book information as HTML:
```
gq -file=sample.html -query="Find .book|Each{OuterHtml}"
```

13. Get the text content of all paragraphs in the document:
```
gq -file=sample.html -query="Find p|Each{Text}"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.
