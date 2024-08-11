package restaurants

import (
	"errors"
	"fmt"
	"log"
	"menucko/services/httpclient"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

const erikaLogPrefix = "[Erika]"
const erikaURL = "https://www.bowlingerika.sk/"
const erikaPDF = "erika.pdf"
const erikaTXT = "erika.txt"

type ErikaParser struct {
	httpClient httpclient.HTTPClient
}

func ParseErika(menuChan chan Menu, waitGroup *sync.WaitGroup, httpClient httpclient.HTTPClient) {
	parser := ErikaParser{
		httpClient: httpClient,
	}

	defer func() {
		if rec := recover(); rec == nil {
			return
		}

		parser.log("Recovered from panic:\n%s", debug.Stack())

		menuChan <- Menu{
			Restaurant: Erika,
			Meals:      nil,
		}
	}()

	defer waitGroup.Done()

	meals, err := parser.parseMenu()
	if err != nil {
		parser.log("Err: %v", err)

		menuChan <- Menu{
			Restaurant: Erika,
			Meals:      nil,
		}
		return
	}

	menuChan <- Menu{
		Restaurant: Erika,
		Meals:      meals,
	}
}

func (parser ErikaParser) parseMenu() (*[]Meal, error) {
	parser.log("Downloading HTML from URL \"%s\"", erikaURL)

	htmlContent, err := parser.httpClient.DownloadHTML(erikaURL)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing HTML from a string with length %d", len(htmlContent))

	rootNode, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	menuPdfLinkSelector := css.MustParse("#denne-menu .elementor-button-link")

	parser.log("Selecting daily menu PDF anchor element with CSS selector")

	menuPdfLinkEls := menuPdfLinkSelector.Select(rootNode)
	if len(menuPdfLinkEls) == 0 {
		return nil, errors.New("daily menu PDF anchor selector didn't match any element")
	}

	parser.log("Selecting \"href\" attribute from the \"a\" daily menu element")
	var aHref string
	for _, imgAttribute := range menuPdfLinkEls[0].Attr {
		if imgAttribute.Key == "href" {
			aHref = imgAttribute.Val
			break
		}
	}

	if len(aHref) == 0 {
		return nil, errors.New("\"a\" daily menu element has no \"href\" attribute")
	}

	parser.log("Downloading PDF from URL \"%s\"", aHref)

	pdfContent, err := parser.httpClient.DownloadHTML(aHref)
	if err != nil {
		return nil, err
	}

	defer parser.cleanUpFiles()

	parser.log("Saving PDF for parsing")

	err = parser.savePDF(&pdfContent)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing PDF content")

	menuText, err := parser.parsePDF()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(menuText, "\n")

	parser.log("Parsing %d lines into individual meals", len(lines))

	meals := make([]Meal, 0)
	var currMeal *Meal
	var mealLength int

	for _, line := range lines {
		cleanedLine := strings.TrimSpace(line)

		if len(cleanedLine) == 0 {
			continue
		}

		if currMeal == nil {
			name, dish := parser.parseFirstLine(cleanedLine)

			if len(name) == 0 {
				continue
			}

			parser.log("Found first line for menu \"%s\"", name)

			meals = append(meals, Meal{
				Name:   name,
				Price:  "",
				Dishes: []string{dish},
			})

			currMeal = &meals[len(meals)-1]

			mealLength = 0

			continue
		}

		price := parser.parsePrice(cleanedLine)
		if len(price) != 0 {
			currMeal.Price = price

			currMeal = nil

			continue
		}

		if mealLength > 4 {
			break
		}

		currMeal.Dishes = append(currMeal.Dishes, cleanedLine)

		mealLength++

	}

	return &meals, nil
}

func (parser ErikaParser) savePDF(content *string) error {
	file, err := os.Create(erikaPDF)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = file.WriteString(*content)

	return err
}

func (parser ErikaParser) parsePDF() (string, error) {
	cmd := exec.Command("pdftotext", erikaPDF, erikaTXT)

	_, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fileContent, err := os.ReadFile(erikaTXT)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}

func (parser ErikaParser) cleanUpFiles() {
	parser.log("Cleaning up files")

	if _, err := os.Stat(erikaPDF); err == nil {
		_ = os.Remove(erikaPDF)
	}

	if _, err := os.Stat(erikaTXT); err == nil {
		_ = os.Remove(erikaTXT)
	}
}

func (parser ErikaParser) parseFirstLine(line string) (string, string) {
	re := regexp.MustCompile(`(M\d+):\s*(.+)$`)

	match := re.FindStringSubmatch(line)
	if match == nil {
		return "", ""
	}

	name := strings.TrimSpace(match[1])
	dish := strings.TrimSpace(match[2])

	return name, dish
}

func (parser ErikaParser) parsePrice(line string) string {
	if !strings.Contains(line, "€") {
		return ""
	}

	price := strings.ReplaceAll(line, ".", ",")

	return price

}

func (parser ErikaParser) log(format string, v ...any) {
	message := erikaLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}
