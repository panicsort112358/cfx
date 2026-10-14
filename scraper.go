package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func scrapeSampleCases(url string) ([]SampleCase, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CFX/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP erro %d when fetching problem page", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var samples []SampleCase
	var traverseHTML func(*html.Node)
	traverseHTML = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "sample-test" {
					if len(samples) == 0 {
						samples = parseSampleTestBlock(n)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverseHTML(c)
		}
	}

	traverseHTML(doc)
	return samples, nil
}

func parseSampleTestBlock(sampleTestNode *html.Node) []SampleCase {
	var inputs []string
	var outputs []string

	var collectAll func(*html.Node)
	collectAll = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					if attr.Val == "input" {
						inputs = append(inputs, extractPreText(n))
					} else if attr.Val == "output" {
						outputs = append(outputs, extractPreText(n))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectAll(c)
		}
	}

	collectAll(sampleTestNode)

	var cases []SampleCase
	limit := len(inputs)
	if len(outputs) < limit {
		limit = len(outputs)
	}

	for i := 0; i < limit; i++ {
		inText := "\n" + strings.TrimSpace(inputs[i])
		outText := "\n" + strings.TrimSpace(outputs[i])
		cases = append(cases, SampleCase{
			Input:  inText,
			Output: outText,
		})
	}

	return cases
}

func extractPreText(n *html.Node) string {
	var textBuilder strings.Builder

	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "br" {
			textBuilder.WriteString("\n")
			return
		}
		if node.Type == html.TextNode {
			textBuilder.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	var findPreTag func(*html.Node) *html.Node
	findPreTag = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode && node.Data == "pre" {
			return node
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if res := findPreTag(c); res != nil {
				return res
			}
		}
		return nil
	}

	if pre := findPreTag(n); pre != nil {
		extract(pre)
	}
	return textBuilder.String()
}
