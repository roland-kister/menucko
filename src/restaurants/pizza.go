package restaurants

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"menucko/services/dateresolver"
	"menucko/services/httpclient"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

const pizzaLogPrefix = "[Pizza]"
const pizzaURL = "https://www.pizza-pizza.sk/menu---terasa"

type PizzaParser struct {
	dateResolver dateresolver.DateResolver
	httpClient   httpclient.HTTPClient
}

func ParsePizza(menuChan chan Menu, waitGroup *sync.WaitGroup, dateResolver dateresolver.DateResolver, httpClient httpclient.HTTPClient) {
	parser := PizzaParser{
		dateResolver: dateResolver,
		httpClient:   httpClient,
	}

	defer func() {
		if rec := recover(); rec == nil {
			return
		}

		parser.log("Recovered from panic:\n%s", debug.Stack())

		menuChan <- Menu{
			Restaurant: Pizza,
			Meals:      nil,
		}
	}()

	defer waitGroup.Done()

	meals, err := parser.parseMenu()
	if err != nil {
		parser.log("Err: %v", err)

		menuChan <- Menu{
			Restaurant: Pizza,
			Meals:      nil,
		}
		return
	}

	menuChan <- Menu{
		Restaurant: Pizza,
		Meals:      meals,
	}
}

func (parser PizzaParser) parseMenu() (*[]Meal, error) {
	parser.log("Downloading HTML from URL \"%s\"", pizzaURL)

	htmlContent, err := parser.httpClient.DownloadHTML(pizzaURL)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing HTML from a string with length %d", len(htmlContent))

	rootNode, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	menuSelectorStr := fmt.Sprintf("#ObedoveMenuu .menuCategory:nth-of-type(%d)", parser.dateResolver.Weekday()+1)
	menuSelector := css.MustParse(menuSelectorStr)

	parser.log("Selecting daily menu element with CSS selector \"%s\"", menuSelectorStr)

	menuEls := menuSelector.Select(rootNode)
	if len(menuEls) == 0 {
		return nil, fmt.Errorf("daily menu CSS selector \"%s\" didn't match any element", menuSelectorStr)
	}

	mealsSelector := css.MustParse(".menuItemBox")

	mealEls := mealsSelector.Select(menuEls[0])
	if len(mealEls) == 0 {
		return nil, errors.New("CSS selector for meals didn't match any element")
	}

	parser.log("Parsing %d meals elements", len(mealEls))

	meals := make([]Meal, 0)

	for index, mealEl := range mealEls {
		name := parser.parseMealName(mealEl)
		price := parser.parseMealPrice(mealEl)
		dishLines := parser.parseDishLines(mealEl)

		if len(name) == 0 || len(price) == 0 || len(dishLines) == 0 {
			parser.log("Meal with index %d has no name, price or dishes", index)
			continue
		}

		parser.log("Parsed meal with name \"%s\"", name)

		meals = append(meals, Meal{
			Name:   name,
			Price:  price,
			Dishes: dishLines,
		})
	}

	return &meals, nil
}

func (parser PizzaParser) parseMealName(mealNode *html.Node) string {
	nameSelector := css.MustParse(".menuItemName")

	nameEls := nameSelector.Select(mealNode)

	if len(nameEls) == 0 {
		return ""
	}

	name := &bytes.Buffer{}
	collectHTMLText(nameEls[0], name)

	return strings.TrimSpace(name.String())
}

func (parser PizzaParser) parseMealPrice(mealNode *html.Node) string {
	priceSelector := css.MustParse(".menuItemPrice")

	priceEls := priceSelector.Select(mealNode)

	if len(priceEls) < 1 {
		return ""
	}

	priceBuf := &bytes.Buffer{}
	collectHTMLText(priceEls[0], priceBuf)

	price := strings.TrimSpace(priceBuf.String())

	if len(price) <= 0 {
		return ""
	}

	price = strings.ReplaceAll(price, ".", ",")

	if strings.Contains(price, "€") {
		return price
	}

	price = price + "€"
	return price
}

func (parser PizzaParser) parseDishLines(mealNode *html.Node) []string {
	dishLineSelector := css.MustParse(".rteBlock")

	dishLineEls := dishLineSelector.Select(mealNode)

	if len(dishLineEls) < 1 {
		return []string{}

	}

	var dishLines []string

	for _, dishLineEl := range dishLineEls {
		dishLineBuffer := &bytes.Buffer{}
		collectHTMLText(dishLineEl, dishLineBuffer)

		dishLine := strings.TrimSpace(dishLineBuffer.String())

		if len(dishLine) == 0 {
			continue
		}

		dishLines = append(dishLines, dishLine)
	}

	return dishLines
}

func (parser PizzaParser) log(format string, v ...any) {
	message := pizzaLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}
