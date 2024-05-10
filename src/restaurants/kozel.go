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

const kozelLogPrefix = "[Kozel]"
const kozelURL = "http://kozeltankpub.sk/obedove-menu/"

type KozelParser struct {
	dateResolver dateresolver.DateResolver
	httpClient   httpclient.HTTPClient
}

func ParseKozel(menuChan chan Menu, waitGroup *sync.WaitGroup, dateResolver dateresolver.DateResolver, httpClient httpclient.HTTPClient) {
	parser := KozelParser{
		dateResolver: dateResolver,
		httpClient:   httpClient,
	}

	defer func() {
		if rec := recover(); rec == nil {
			return
		}

		parser.log("Recovered from panic:\n%s", debug.Stack())

		menuChan <- Menu{
			Restaurant: Kozel,
			Meals:      nil,
		}
	}()

	defer waitGroup.Done()

	meals, err := parser.parseMenu()
	if err != nil {
		parser.log("Err: %v", err)

		menuChan <- Menu{
			Restaurant: Kozel,
			Meals:      nil,
		}
		return
	}

	menuChan <- Menu{
		Restaurant: Kozel,
		Meals:      meals,
	}
}

func (parser KozelParser) parseMenu() (*[]Meal, error) {
	parser.log("Downloading HTML from URL \"%s\"", kozelURL)

	htmlContent, err := parser.httpClient.DownloadHTML(kozelURL)
	if err != nil {
		return nil, err
	}

	parser.log("Parsing HTML from a string with length %d", len(htmlContent))

	rootNode, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	menuEl, err := parser.findDailyMenuEl(rootNode)
	if err != nil {
		return nil, err
	}

	meals := make([]Meal, 0)

	parser.log("Parsing soups")

	soupsSelector := css.MustParse(".polievky .menu-holder")

	soupsEls := soupsSelector.Select(menuEl)
	if len(soupsEls) > 0 {
		soupMeal := parser.parseSoups(soupsEls)
		meals = append(meals, soupMeal)
	}

	parser.log("Parsing main meals")

	mainMealsSelector := css.MustParse(".hlavne .menu-holder")

	mainMealsEls := mainMealsSelector.Select(menuEl)
	if len(mainMealsEls) == 0 {
		return nil, errors.New("CSS selector for meals didn't match any element")
	}

	for index, mainMealEl := range mainMealsEls {
		name := parser.parseMealName(mainMealEl)
		price := parser.parseMealPrice(mainMealEl)
		dish := parser.parseDishName(mainMealEl)

		if len(name) == 0 || len(price) == 0 || len(dish) == 0 {
			parser.log("Meal with index %d has no name, price or dish", index)
			continue
		}

		parser.log("Parsed meal with name \"%s\"", name)

		meals = append(meals, Meal{
			Name:   name,
			Price:  price,
			Dishes: []string{dish},
		})
	}

	return &meals, nil
}

func (parser KozelParser) findDailyMenuEl(rootNode *html.Node) (*html.Node, error) {
	parser.log("Selecting daily menu elements")

	menuSelector := css.MustParse(".entry-content .daily-menu")

	menuEls := menuSelector.Select(rootNode)
	if len(menuEls) == 0 {
		return nil, errors.New("daily menu CSS selector didn't match any element")
	}

	kozelDayName := strings.ToLower(parser.dateResolver.SlovakWeekday())

	parser.log("Looking for today's daily menu element for day \"%s\"", kozelDayName)

	daySelector := css.MustParse("h3")

	for _, menuEl := range menuEls {
		dayEls := daySelector.Select(menuEl)

		if len(dayEls) == 0 {
			continue
		}

		dayNameBuf := &bytes.Buffer{}
		collectHTMLText(dayEls[0], dayNameBuf)

		dayName := strings.TrimSpace(dayNameBuf.String())
		dayName = strings.ToLower(dayName)

		if strings.HasPrefix(dayName, kozelDayName) {
			return menuEl, nil
		}
	}

	return nil, errors.New("daily menu CSS selector didn't match any element")
}

func (parser KozelParser) parseSoups(soupsEls []*html.Node) Meal {
	name := parser.parseMealName(soupsEls[0])

	var soupNames []string

	for _, soupEl := range soupsEls {
		soupName := parser.parseDishName(soupEl)

		if len(soupName) > 0 {
			soupNames = append(soupNames, soupName)
		}
	}

	return Meal{
		Name:   name,
		Price:  "",
		Dishes: soupNames,
	}
}

func (parser KozelParser) parseMealName(mealEl *html.Node) string {
	nameSelector := css.MustParse("span:first-of-type")

	nameEls := nameSelector.Select(mealEl)
	if len(nameEls) == 0 {
		return ""
	}

	name := &bytes.Buffer{}
	collectHTMLText(nameEls[0], name)

	return strings.TrimSpace(name.String())
}

func (parser KozelParser) parseDishName(mealEl *html.Node) string {
	dishSelector := css.MustParse("p")

	dishEls := dishSelector.Select(mealEl)
	if len(dishEls) == 0 {
		return ""
	}

	dish := &bytes.Buffer{}
	collectHTMLText(dishEls[0], dish)

	return strings.TrimSpace(dish.String())
}

func (parser KozelParser) parseMealPrice(mealEl *html.Node) string {
	priceSelector := css.MustParse(".menu-price")

	priceEls := priceSelector.Select(mealEl)

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

func (parser KozelParser) log(format string, v ...any) {
	message := kozelLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}
