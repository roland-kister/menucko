package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"menucko/restaurants"
	"time"
	_ "time/tzdata"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

const htmlRendererLogPrefix = "[HTML Renderer]"
const rendererFatalErrPage = "<!doctype html><html lang=sk><h1>Fatal Error</h1>"

var slovakDays = [...]string{"Nedeľa", "Pondelok", "Utorok", "Streda", "Štvrtok", "Piatok", "Sobota"}

type Renderer interface {
	RenderMenus(menus *[]restaurants.Menu) (*bytes.Buffer, error)
	GetErrorContent() *bytes.Buffer
}

type HTMLRenderer struct {
	TemplateFilePath string
	StylesPath       string
	CommitHash       string
}

type HTMLRendererContent struct {
	Menus         *[]restaurants.Menu
	StylesPath    string
	CommitHash    string
	ExecutionTime string
	DayName       string
}

func (r HTMLRenderer) RenderMenus(menus *[]restaurants.Menu) (*bytes.Buffer, error) {
	r.log("Loading HTML template from \"%s\"", r.TemplateFilePath)

	temp, err := template.ParseFiles(r.TemplateFilePath)
	if err != nil {
		r.err(err)
		return nil, err
	}

	loc, err := time.LoadLocation("Europe/Bratislava")
	if err != nil {
		r.err(err)
		return nil, err
	}

	currentTime := time.Now().In(loc)

	content := HTMLRendererContent{
		Menus:         menus,
		StylesPath:    r.StylesPath,
		CommitHash:    r.CommitHash,
		ExecutionTime: currentTime.Format("15:04 2.1.2006"),
		DayName:       slovakDays[currentTime.Weekday()],
	}

	r.log("Rendering HTML content")
	renderBuff := new(bytes.Buffer)

	err = temp.Execute(renderBuff, content)
	if err != nil {
		r.err(err)
		return nil, err
	}

	r.log("Minifying HTML content")

	minifier := minify.New()
	minifier.AddFunc("text/html", html.Minify)

	minifyBuff := new(bytes.Buffer)

	err = minifier.Minify("text/html", minifyBuff, renderBuff)
	if err != nil {
		r.err(err)
		return nil, err
	}

	return minifyBuff, nil
}

func (r HTMLRenderer) GetErrorContent() *bytes.Buffer {
	buff := new(bytes.Buffer)

	_, _ = fmt.Fprint(buff, rendererFatalErrPage)

	return buff
}

func (HTMLRenderer) log(format string, v ...any) {
	message := htmlRendererLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}

func (HTMLRenderer) err(err error) {
	message := htmlRendererLogPrefix + fmt.Sprintf(" Err: %v", err)

	log.Println(message)
}
