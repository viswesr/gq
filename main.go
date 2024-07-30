package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	fileFlag := flag.String("file", "", "HTML file to parse")
	queryFlag := flag.String("query", "", "Query to execute")
	flag.Parse()

	if *fileFlag == "" || *queryFlag == "" {
		fmt.Println("Usage: go run main.go -file=<html_file> -query=<query>")
		os.Exit(1)
	}

	file, err := os.Open(*fileFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
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

// trimSpaceAndNewline from beginning and end of the string
func trimSpaceAndNewline(s string) string {
	return strings.Trim(s, " \n")
}
