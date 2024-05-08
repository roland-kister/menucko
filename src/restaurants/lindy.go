package restaurants

import (
	"errors"
	"fmt"
	"log"
	"menucko/util"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

const lindyLogPrefix = "[Lindy]"
const lindyURL = "http://www.lindyhop.sk/"

type LindyParser struct {
	httpClient util.HTTPClient
	imageOcr   util.ImageOcr
}

func ParseLindy(menuChan chan Menu, waitGroup *sync.WaitGroup, httpClient util.HTTPClient, imageOcr util.ImageOcr) {
	parser := LindyParser{
		httpClient: httpClient,
		imageOcr:   imageOcr,
	}

	defer func() {
		if rec := recover(); rec == nil {
			return
		}

		parser.log("Recovered from panic:\n%s", debug.Stack())

		menuChan <- Menu{
			Restaurant: Lindy,
			Meals:      nil,
		}
	}()

	defer waitGroup.Done()

	meals, err := parser.parseMenu()
	if err != nil {
		parser.log("Err: %v", err)

		menuChan <- Menu{
			Restaurant: Lindy,
			Meals:      nil,
		}
		return
	}

	menuChan <- Menu{
		Restaurant: Lindy,
		Meals:      meals,
	}
}

func (parser LindyParser) parseMenu() (*[]Meal, error) {
	parser.log("Downloading HTML from URL \"%s\"", lindyURL)

	htmlContent, err := parser.httpClient.DownloadHTML(lindyURL)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing HTML from a string with length %d", len(htmlContent))

	rootNode, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	menuImgSelector := css.MustParse("#DenneMenu img")

	parser.log("Selecting daily menu image element with CSS selector")

	menuImgEls := menuImgSelector.Select(rootNode)
	if len(menuImgEls) == 0 {
		return nil, errors.New("daily menu image CSS selector didn't match any element")
	}

	parser.log("Selecting \"src\" attribute from the \"img\" daily menu element")

	var imgSrc string
	for _, imgAttribute := range menuImgEls[0].Attr {
		if imgAttribute.Key == "src" {
			imgSrc = imgAttribute.Val
			break
		}
	}

	if len(imgSrc) == 0 {
		return nil, errors.New("\"img\" daily menu element has no \"src\" attribute")
	}

	imageURL := lindyURL + imgSrc

	parser.log("Downloading menu image from URL \"%s\"", imageURL)

	imageContent, err := parser.httpClient.DownloadHTML(imageURL)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing text from image with length %d", len(imageContent))

	imgText, err := parser.imageOcr.ParseJpegText([]byte(imageContent))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(imgText, "\n")

	parser.log("Parsing %d lines into individual meals", len(lines))

	meals := make([]Meal, 0)
	var currMeal *Meal
	var mealLength int

	for _, line := range lines {
		cleanedLine := strings.TrimSpace(line)

		if len(cleanedLine) == 0 {
			continue
		}

		if strings.HasSuffix(cleanedLine, "€") {
			name, price := parser.parseNamePrice(cleanedLine)

			parser.log("Found first line for menu \"%s\"", name)

			meals = append(meals, Meal{
				Name:   name,
				Price:  price,
				Dishes: []string{},
			})

			currMeal = &meals[len(meals)-1]

			mealLength = 0

			continue
		}

		if (currMeal != nil && strings.Contains(cleanedLine, "Ponuka ")) || mealLength > 4 {
			break
		}

		if currMeal != nil {
			allergensRe := regexp.MustCompile(` [^ ]*\d[^ ]*$`)
			cleanedLine = allergensRe.ReplaceAllString(cleanedLine, "")

			currMeal.Dishes = append(currMeal.Dishes, cleanedLine)

			mealLength++
		}
	}

	return &meals, nil
}

func (parser LindyParser) parseNamePrice(line string) (string, string) {
	cleanedLine := strings.ReplaceAll(line, ".", ",")
	namePriceRe := regexp.MustCompile(`(?U)(.+) (\d+,\d+)\s+€`)

	match := namePriceRe.FindStringSubmatch(cleanedLine)
	if match == nil {
		return "", ""
	}

	name := strings.TrimSpace(match[1])
	price := match[2] + "€"

	return name, price
}

func (parser LindyParser) log(format string, v ...any) {
	message := lindyLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}
