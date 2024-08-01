package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yosssi/gohtml"
)

func main() {
	fileFlag := flag.String("file", "", "HTML file to parse")
	urlFlag := flag.String("url", "", "HTML URL to download and parse")
	queryFlag := flag.String("query", "", "Query to execute")
	codeGenFlag := flag.Bool("gencode", false, "Generate Go code snippet for the query")
	flag.Parse()

	if *codeGenFlag {
		generateGoCode(*fileFlag, *queryFlag)
		return
	}

	if (*fileFlag == "" && *urlFlag == "") || *queryFlag == "" {
		fmt.Println("Usage:\n gq -file=<html_file> -query=<query eg:Find a|Each{Attrib href}> or\n gq -url=<html_url> -query=<query eg:Find a|Each{Attrib href}>\nThere is an optional `-gencode` flag that generates `go` code for query:\n gq -file=<html_file> -query=<query> -gencode")
		os.Exit(0)
	} else if *fileFlag != "" && *urlFlag != "" {
		fmt.Println("Please provide either a -file or -url, not both")
		os.Exit(0)
	}

	*fileFlag = *urlFlag

	var doc *goquery.Document
	if strings.HasPrefix(*fileFlag, "http:") || strings.HasPrefix(*fileFlag, "https:") {
		resp, err := http.Get(*fileFlag)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		file, err := os.Open(*fileFlag)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		doc, err = goquery.NewDocumentFromReader(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Parse and execute the query
	executeQuery(doc.Selection, *queryFlag)
}

func executeQuery(selection *goquery.Selection, query string) {
	parts := parseParts(query)

	for _, part := range parts {
		if strings.HasPrefix(part, "Find ") {
			selector := strings.TrimPrefix(part, "Find ")
			selection = selection.Find(selector)
		} else if strings.HasPrefix(part, "Each{") {
			subQuery := strings.TrimPrefix(part, "Each{")
			subQuery = strings.TrimSuffix(subQuery, "}")
			selection.Each(func(i int, s *goquery.Selection) {
				fmt.Printf("%d: ", i+1)
				executeQuery(s, subQuery)
			})
			return
		} else if part == "First" {
			selection = selection.First()
		} else if part == "Last" {
			selection = selection.Last()
		} else if part == "Parent" {
			selection = selection.Parent()
		} else if part == "Children" {
			selection = selection.Children()
		} else if strings.HasPrefix(part, "Attrib ") {
			attr := strings.TrimPrefix(part, "Attrib ")
			val, exists := selection.Attr(attr)
			if exists {
				printColourized(val, "markdown")
			}
			return
		} else if part == "Text" {
			printColourized(selection.Text(), "markdown")
			return
		} else if part == "Html" {
			html, err := selection.Html()
			if err == nil {
				printColourized(trimSpaceAndNewline(html), "html")
			}
			return
		} else if part == "OuterHtml" {
			html, err := goquery.OuterHtml(selection)
			if err == nil {
				printColourized(trimSpaceAndNewline(html), "html")
			}
			return
		}
	}
}

func parseParts(query string) []string {
	var parts []string
	var currentPart strings.Builder
	depth := 0

	for _, char := range query {
		switch char {
		case '{':
			depth++
			currentPart.WriteRune(char)
		case '}':
			depth--
			currentPart.WriteRune(char)
			if depth == 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		case '|':
			if depth == 0 {
				if currentPart.Len() > 0 {
					parts = append(parts, currentPart.String())
					currentPart.Reset()
				}
			} else {
				currentPart.WriteRune(char)
			}
		default:
			currentPart.WriteRune(char)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

func generateGoCode(file, query string) {
	isHTTP := strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:")
	importPackage := "os"
	if isHTTP {
		importPackage = "net/http"
	}
	code := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	%q
	"github.com/PuerkitoBio/goquery"
)

func main() {
	fileName := %q
	var doc *goquery.Document
	var err error

`, importPackage, file)

	if isHTTP {
		code += `	resp, err := http.Get(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, err = goquery.NewDocumentFromReader(resp.Body)
`
	} else {
		code += `	file, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	doc, err = goquery.NewDocumentFromReader(file)
`
	}

	code += `	if err != nil {
		log.Fatal(err)
	}

`
	code += generateQueryCode("\t", "doc", query)
	code += "}"
	printColourized(code, "go")
}

func generateQueryCode(indent string, selection string, query string) string {
	var code strings.Builder
	parts := parseParts(query)

	for i, part := range parts {
		switch {
		case strings.HasPrefix(part, "Find "):
			selector := strings.TrimPrefix(part, "Find ")
			selection = fmt.Sprintf("%s.Find(%q)", selection, selector)
		case strings.HasPrefix(part, "Each{"):
			subQuery := strings.TrimPrefix(part, "Each{")
			subQuery = strings.TrimSuffix(subQuery, "}")
			code.WriteString(fmt.Sprintf("%s%s.Each(func(i int, s *goquery.Selection) {\n", indent, selection))
			code.WriteString(fmt.Sprintf("%s\tfmt.Printf(\"%%d: \", i+1)\n", indent))
			code.WriteString(generateQueryCode(indent+"\t", "s", subQuery))
			code.WriteString(fmt.Sprintf("%s})\n", indent))
			return code.String()
		case part == "First":
			selection = fmt.Sprintf("%s.First()", selection)
		case part == "Last":
			selection = fmt.Sprintf("%s.Last()", selection)
		case part == "Parent":
			selection = fmt.Sprintf("%s.Parent()", selection)
		case part == "Children":
			selection = fmt.Sprintf("%s.Children()", selection)
		case strings.HasPrefix(part, "Attrib "):
			attr := strings.TrimPrefix(part, "Attrib ")
			code.WriteString(fmt.Sprintf("%sval, exists := %s.Attr(%q)\n", indent, selection, attr))
			code.WriteString(fmt.Sprintf("%sif exists {\n", indent))
			code.WriteString(fmt.Sprintf("%s\tfmt.Println(val)\n", indent))
			code.WriteString(fmt.Sprintf("%s}\n", indent))
			return code.String()
		case part == "Text":
			code.WriteString(fmt.Sprintf("%sfmt.Println(%s.Text())\n", indent, selection))
			return code.String()
		case part == "Html":
			code.WriteString(fmt.Sprintf("%shtml, err := %s.Html()\n", indent, selection))
			code.WriteString(fmt.Sprintf("%sif err == nil {\n", indent))
			code.WriteString(fmt.Sprintf("%s\tfmt.Println(strings.TrimSpace(html))\n", indent))
			code.WriteString(fmt.Sprintf("%s}\n", indent))
			return code.String()
		case part == "OuterHtml":
			code.WriteString(fmt.Sprintf("%shtml, err := goquery.OuterHtml(%s)\n", indent, selection))
			code.WriteString(fmt.Sprintf("%sif err == nil {\n", indent))
			code.WriteString(fmt.Sprintf("%s\tfmt.Println(strings.TrimSpace(html))\n", indent))
			code.WriteString(fmt.Sprintf("%s}\n", indent))
			return code.String()
		}

		if i == len(parts)-1 {
			code.WriteString(fmt.Sprintf("%sfmt.Println(%s)\n", indent, selection))
		}
	}

	return code.String()
}

// printColourized prints the text or HTML or Go snippet with syntax highlighting
func printColourized(h string, t string) {
	gohtml.Condense = true
	beautifiedHTML := gohtml.Format(h)
	lexer := lexers.Get(t)
	style := styles.Get("fruity")
	formatter := formatters.Get("terminal256")
	iterator, err := lexer.Tokenise(nil, beautifiedHTML)
	if err != nil {
		fmt.Println("Error tokenizing HTML:", err)
		return
	}
	err = formatter.Format(os.Stdout, style, iterator)
	if err != nil {
		fmt.Println("Error formatting HTML:", err)
		return
	}
	fmt.Println("")
}

// trimSpaceAndNewline from beginning and end of the string
func trimSpaceAndNewline(s string) string {
	return strings.Trim(s, " \n")
}
