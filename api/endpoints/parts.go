package endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"net/http"
	"pcpartpicker-api/api/entities"
	"pcpartpicker-api/cache"
	"pcpartpicker-api/scraper"
	"sync"
	"time"
)

func GetPartsList(w http.ResponseWriter, r *http.Request) {
	path := r.Header.Get("path")

	parts, cached := GetParts(path)

	_ = json.NewEncoder(w).Encode(parts)

	if !cached {
		b, _ := json.Marshal(parts)
		cache.Put(path, b)
	}
}

//Returns true if value is retrieved from Cache, else false
func GetParts(path string) (entities.Parts, bool) {
	if data, success := cache.RetrieveCache(path); success {
		var db entities.Parts

		_ = json.Unmarshal(data, &db)

		return db, true
	}

	if _, err := scraper.Instance.ExecuteScript(fmt.Sprintf("window.open('%s');", path), nil); err != nil {
		log.Println(err)
	}

	windows, _ := scraper.Instance.WindowHandles()
	_ = scraper.Instance.SwitchWindow(windows[len(windows)-1])

	handle, _ := scraper.Instance.CurrentWindowHandle()
	defer scraper.Instance.SwitchWindow(windows[0])
	defer scraper.Instance.CloseWindow(handle)

	var parts entities.Parts

	err := scraper.Instance.WaitWithTimeout(func(wd selenium.WebDriver) (b bool, err error) {
		wrapper, _ := wd.FindElements(selenium.ByCSSSelector, ".tr__product")
		if len(wrapper) > 0 {
			return true, nil
		}

		return false, nil
	}, time.Minute)

	if err != nil {
		log.Println(err)
	}

	compat, _ := scraper.Instance.FindElement(selenium.ByCSSSelector, ".partlist__metrics p")
	compatP, _ := compat.GetAttribute("class")

	if compatP == "partlist__compatibility--noIssues" {
		parts.Compatibility = true
	}

	watt, _ := scraper.Instance.FindElement(selenium.ByCSSSelector, ".partlist__keyMetric a")
	parts.Wattage, _ = watt.Text()

	components, _ := scraper.Instance.FindElements(selenium.ByCSSSelector, ".tr__product")

	var wg sync.WaitGroup
	for _, comp := range components {
		wg.Add(1)
		go appendComponents(comp, &parts, &wg)
	}

	wg.Wait()

	return parts, false
}

func appendComponents(comp selenium.WebElement, parts *entities.Parts, wg *sync.WaitGroup) {
	defer wg.Done()

	c, _ := comp.FindElement(selenium.ByCSSSelector, ".td__component a")
	cName, _ := c.Text()

	image, _ := comp.FindElement(selenium.ByCSSSelector, ".td__image img")
	src, _ := image.GetAttribute("src")

	name, _ := comp.FindElement(selenium.ByCSSSelector, ".td__name a")
	nameText, _ := name.Text()

	var priceText string
	price, err := comp.FindElement(selenium.ByCSSSelector, ".td__price a")
	if err != nil  {
		priceText = "No Prices Available"
	} else {
		priceText, _ = price.Text()
	}

	var whereText string
	where, err := comp.FindElement(selenium.ByCSSSelector, ".td__where a")
	if err != nil {
		whereText = "Not Available at any Place"
	} else {
		whereText, _ = where.GetAttribute("href")
	}

	switch cName {
	case "CPU":
		parts.CPU.Title = nameText
		parts.CPU.Image = src
		parts.CPU.Price = priceText
		parts.CPU.Where = whereText
	case "Motherboard":
		parts.Motherboard.Title = nameText
		parts.Motherboard.Image = src
		parts.Motherboard.Price = priceText
		parts.Motherboard.Where = whereText
	case "Memory":
		parts.Memory.Title = nameText
		parts.Memory.Image = src
		parts.Memory.Price = priceText
		parts.Memory.Where = whereText
	case "Storage":
		parts.Storage.Title = nameText
		parts.Storage.Image = src
		parts.Storage.Price = priceText
		parts.Storage.Where = whereText
	case "Video Card":
		parts.VideoCard.Title = nameText
		parts.VideoCard.Image = src
		parts.VideoCard.Price = priceText
		parts.VideoCard.Where = whereText
	case "Case":
		parts.Case.Title = nameText
		parts.Case.Image = src
		parts.Case.Price = priceText
		parts.Case.Where = whereText
	case "Power Supply":
		parts.PSU.Title = nameText
		parts.PSU.Image = src
		parts.PSU.Price = priceText
		parts.PSU.Where = whereText
	}
}
