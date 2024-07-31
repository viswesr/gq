package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	fileFlag := flag.String("file", "", "HTML file to parse")
	queryFlag := flag.String("query", "", "Query to execute")
	codeGenFlag := flag.Bool("gencode", false, "Generate Go code snippet for the query")
	flag.Parse()

	if *codeGenFlag {
		generateGoCode(*fileFlag, *queryFlag)
		return
	}

	if *fileFlag == "" || *queryFlag == "" {
		fmt.Println("Usage: gq -file=<html_file> -query=<query eg:Find a|Each{Attrib href}>")
		os.Exit(1)
	}

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
				fmt.Println(val)
			}
			return
		} else if part == "Text" {
			fmt.Println(selection.Text())
			return
		} else if part == "Html" {
			html, err := selection.Html()
			if err == nil {
				fmt.Println(trimSpaceAndNewline(html))
			}
			return
		} else if part == "OuterHtml" {
			html, err := goquery.OuterHtml(selection)
			if err == nil {
				fmt.Println(trimSpaceAndNewline(html))
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
	fmt.Println("package main")
	fmt.Println()
	fmt.Println("import (")
	fmt.Println("\t\"fmt\"")
	fmt.Println("\t\"log\"")
	if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
		fmt.Println("\t\"net/http\"")
	} else {
		fmt.Println("\t\"os\"")
	}
	fmt.Println("\t\"github.com/PuerkitoBio/goquery\"")
	fmt.Println(")")
	fmt.Println()
	fmt.Println("func main() {")
	fmt.Printf("\tfile := %q\n", file)
	fmt.Println("\tvar doc *goquery.Document")
	fmt.Println("\tvar err error")
	fmt.Println()

	if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
		fmt.Println("\tresp, err := http.Get(file)")
		fmt.Println("\tif err != nil {")
		fmt.Println("\t\tlog.Fatal(err)")
		fmt.Println("\t}")
		fmt.Println("\tdefer resp.Body.Close()")
		fmt.Println("\tdoc, err = goquery.NewDocumentFromReader(resp.Body)")
	} else {
		fmt.Println("\tfile, err := os.Open(file)")
		fmt.Println("\tif err != nil {")
		fmt.Println("\t\tlog.Fatal(err)")
		fmt.Println("\t}")
		fmt.Println("\tdefer file.Close()")
		fmt.Println("\tdoc, err = goquery.NewDocumentFromReader(file)")
	}

	fmt.Println("\tif err != nil {")
	fmt.Println("\t\tlog.Fatal(err)")
	fmt.Println("\t}")
	fmt.Println()

	generateQueryCode("\t", "doc.Selection", query)

	fmt.Println("}")
}

func generateQueryCode(indent string, selection string, query string) {
	parts := parseParts(query)

	for i, part := range parts {
		if strings.HasPrefix(part, "Find ") {
			selector := strings.TrimPrefix(part, "Find ")
			selection = fmt.Sprintf("%s.Find(%q)", selection, selector)
		} else if strings.HasPrefix(part, "Each{") {
			subQuery := strings.TrimPrefix(part, "Each{")
			subQuery = strings.TrimSuffix(subQuery, "}")
			fmt.Printf("%s%s.Each(func(i int, s *goquery.Selection) {\n", indent, selection)
			fmt.Printf("%s\tfmt.Printf(\"%%d: \", i+1)\n", indent)
			generateQueryCode(indent+"\t", "s", subQuery)
			fmt.Printf("%s})\n", indent)
			return
		} else if part == "First" {
			selection = fmt.Sprintf("%s.First()", selection)
		} else if part == "Last" {
			selection = fmt.Sprintf("%s.Last()", selection)
		} else if part == "Parent" {
			selection = fmt.Sprintf("%s.Parent()", selection)
		} else if part == "Children" {
			selection = fmt.Sprintf("%s.Children()", selection)
		} else if strings.HasPrefix(part, "Attrib ") {
			attr := strings.TrimPrefix(part, "Attrib ")
			fmt.Printf("%sval, exists := %s.Attr(%q)\n", indent, selection, attr)
			fmt.Printf("%sif exists {\n", indent)
			fmt.Printf("%s\tfmt.Println(val)\n", indent)
			fmt.Printf("%s}\n", indent)
			return
		} else if part == "Text" {
			fmt.Printf("%sfmt.Println(%s.Text())\n", indent, selection)
			return
		} else if part == "Html" {
			fmt.Printf("%shtml, err := %s.Html()\n", indent, selection)
			fmt.Printf("%sif err == nil {\n", indent)
			fmt.Printf("%s\tfmt.Println(strings.TrimSpace(html))\n", indent)
			fmt.Printf("%s}\n", indent)
			return
		} else if part == "OuterHtml" {
			fmt.Printf("%shtml, err := goquery.OuterHtml(%s)\n", indent, selection)
			fmt.Printf("%sif err == nil {\n", indent)
			fmt.Printf("%s\tfmt.Println(strings.TrimSpace(html))\n", indent)
			fmt.Printf("%s}\n", indent)
			return
		}

		if i == len(parts)-1 {
			fmt.Printf("%sfmt.Println(%s)\n", indent, selection)
		}
	}
}

// trimSpaceAndNewline from beginning and end of the string
func trimSpaceAndNewline(s string) string {
	return strings.Trim(s, " \n")
}
