package main

import (
	"log"
	"menucko/restaurants"
	"menucko/services/httpclient"
	"menucko/services/imageocr"
	"sync"
)

func main() {
	dateResolver, err := getDateResolver()
	if err != nil {
		log.Println(err)
		return
	}

	httpClient := httpclient.ProdHTTPClient{}
	imageOcr := imageocr.ProdImageOcr{}

	rend, err := getRenderer()
	if err != nil {
		log.Println(err)
		return
	}

	dist, err := getDistributor()
	if err != nil {
		log.Println(err)
		return
	}

	menuChan := make(chan restaurants.Menu, 4)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go restaurants.ParsePizza(menuChan, &waitGroup, dateResolver, httpClient)

	waitGroup.Add(1)
	go restaurants.ParseKozel(menuChan, &waitGroup, dateResolver, httpClient)

	waitGroup.Add(1)
	go restaurants.ParseLindy(menuChan, &waitGroup, httpClient, imageOcr)

	waitGroup.Add(1)
	go restaurants.ParseErika(menuChan, &waitGroup, httpClient)

	waitGroup.Wait()

	close(menuChan)

	var menus [4]restaurants.Menu
	for menu := range menuChan {
		menus[menu.Restaurant] = menu
	}

	menusSlice := menus[:]

	htmlContent, err := rend.RenderMenus(&menusSlice)
	if err != nil {
		htmlContent = rend.GetErrorContent()
	}

	err = dist.Distribute(htmlContent)
	if err != nil {
		log.Println(err)
	}
}
