package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"regexp"
	"strconv"
	"time"
	"wysh-app/models"
)

var articles []models.ArticleMini
var trends []models.TrendMini
var i = 0

var pageUrls []string

func GetFreakinsData(c_ *gin.Context) {
	pageUrls = append(pageUrls, "https://freakins.com/collections/denim-dresses", "https://freakins.com/collections/denim-tops")
	trendTitles := []string{"Freakins Dresses", "Freakins Tops"}
	defer timeTrack(time.Now(), "Freakins")
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		//colly.MaxDepth(2),
		colly.Async(true),
	)
	//err := c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 6})
	//if err != nil {
	//	return
	//}

	c.OnRequest(onRequestCallback)
	c.OnHTML(".ProductItem__ImageWrapper", visitProductDetailPage)
	//c.OnHTML(".ProductMeta__Title", parseProductDetailPage)
	c.OnHTML("#shopify-section-product-template > section", parseProductDetailPage)
	c.OnHTML("#ProductGridContainer > div.Pagination.Text--subdued > div", visitNextPage)
	for i, pageUrl := range pageUrls {
		err := c.Visit(pageUrl)
		if err != nil {
			return
		}
		c.Wait()
		trends = append(trends, models.TrendMini{
			ID:       i,
			Name:     trendTitles[i],
			Articles: articles,
		})
		articles = []models.ArticleMini{}
	}
	c_.JSON(200, models.HomeResObj{Trends: trends})
}

func visitProductDetailPage(e *colly.HTMLElement) {
	url := e.ChildAttr("a[href]", "href")
	//fmt.Println(e.Request.AbsoluteURL(url))
	err := e.Request.Visit(e.Request.AbsoluteURL(url))
	if err != nil {
		return
	}
}

func visitNextPage(e *colly.HTMLElement) {
	url := e.ChildAttr("a[href]", "href")
	//fmt.Println(e.Request.AbsoluteURL(url))
	err := e.Request.Visit(e.Request.AbsoluteURL(url))
	if err != nil {
		return
	}
}

func parseProductDetailPage(e *colly.HTMLElement) {
	imageUrls := e.ChildAttrs("img", "src")
	imageUrls = imageUrls[:len(imageUrls)-1]
	for i := range imageUrls {
		imageUrls[i] = "http:" + imageUrls[i]
	}
	priceStr := e.ChildText("span.ProductMeta__Price > span.money")
	priceRegex := regexp.MustCompile("[^0-9]")
	price, _ := strconv.Atoi(priceRegex.ReplaceAllString(priceStr[4:], ""))

	articles = append(articles, models.ArticleMini{
		ID:           int64(i),
		Name:         e.ChildText(".ProductMeta__Title"),
		Brand:        "Freakins",
		CurrentPrice: int64(price),
		BasePrice:    int64(price),
		ArticleUrl:   e.Request.AbsoluteURL(e.Request.URL.EscapedPath()),
		ImageUrl:     imageUrls[0],
	})
	i++
}

func onRequestCallback(r *colly.Request) {
	fmt.Println("Visiting", r.URL)
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s", name, elapsed)
}
