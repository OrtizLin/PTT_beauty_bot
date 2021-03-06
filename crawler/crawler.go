package crawler

import (
	"github.com/PuerkitoBio/goquery"
	"line_bot_final/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const BasePttAddress = "https://www.ptt.cc"

type Article struct {
	Title           string
	LikeCount       int
	Link            string
	Date            string
	ImageLink       string
	LikeCountString string
	Board 			string
}

type HotBoard struct {
	Board 			string
}

func Start(w http.ResponseWriter, r *http.Request) {
	db.RemoveALL("xtest")
	
	// 用來抓取最新熱門看板，不用每次都跑，
	// db.RemoveALL("hotboard")
	// getHotBoards() 

	// 撈所有熱門看板的當日文章
	// var results []string = db.AllHotBoards()
	// for i:= 0; i < len(results); i++ {
	// 	go getAllArticles(results[i])
	// }

	// 只撈表特+西斯版
	go getAllArticles("Beauty")
	// go getAllArticles("Sex")
}

func getHotBoards() { // 取得熱門看板

	var url string = BasePttAddress + "/bbs/hotboards.html"
	var boards []string

	client:=&http.Client{}
	req := passR18(url)
	res, err := client.Do(req)
	defer res.Body.Close()

	// 最後直接把res傳给goquery就可以來解析網頁
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find(".b-ent").Each(func(i int, s *goquery.Selection) {
			boards = append(boards, s.Find(".board-name").Text())
		})
		
		db.InsertHotBoard(boards)

}

func getAllArticles(forum string) {

	var BOOL = true
	var exist = true
	var nextURL string = ""  // default url
	var href string = "" // next page url
	var crawlerCount = 0

	// today's date
	loc, _ := time.LoadLocation("Asia/Chongqing")
	time := time.Now().In(loc)

	// 開始爬蟲
	for BOOL {

		if href == "" {
			nextURL = BasePttAddress + "/bbs/" + forum + "/index.html" // 首頁
		} else {
			nextURL = BasePttAddress + href // 翻至下一頁
		}

	client:=&http.Client{}
	req := passR18(nextURL)
	res, err := client.Do(req)
	defer res.Body.Close()

	// 最後直接把res傳给goquery就可以來解析網頁
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			log.Fatal(err)
		}

		//Find previous link
		doc.Find(".btn-group a").Each(func(i int, s *goquery.Selection) {

			if strings.Contains(s.Text(), "上頁") {
				href, exist = s.Attr("href")
			}

		})

		doc.Find(".r-ent").Each(func(i int, s *goquery.Selection) {
			article := Article{}
			article.Title = strings.TrimSpace(s.Find(".title").Text())
			article.LikeCount, _ = strconv.Atoi(s.Find(".nrec span").Text())
			hrefs, _ := s.Find(".title a").Attr("href")
			article.Link = BasePttAddress + hrefs
			article.Date = strings.TrimSpace(s.Find(".meta").Find(".date").Text())
			article.ImageLink = "https://i.imgur.com/wIdGRrU.jpg" // 先塞入預設imageLink
			article.LikeCountString = s.Find(".nrec span").Text()
			article.Board = forum
			if article.Date != time.Format("1/02") {
				if crawlerCount > 0 {
					BOOL = false // 爬不到今日文章後 停止爬蟲
				}
			}

			// 今日文章且未被刪除（被刪除文章url會變成BasePttAddress)
			// 若文章內含有https及.jpg 的字串, 儲存為article.ImageLink.
			if article.Date == time.Format("1/02") && article.Link != BasePttAddress {
				//search image link in article

			client:=&http.Client{}
			req := passR18(article.Link)
			res, err := client.Do(req)
			defer res.Body.Close()

			// 最後直接把res傳给goquery就可以來解析網頁
				doc, err := goquery.NewDocumentFromResponse(res)
				if err != nil {
					log.Fatal(err)
				}

				doc.Find("#main-content > a").EachWithBreak(func(i int, s *goquery.Selection) bool {
					imgLink := s.Text()
					if strings.Contains(imgLink, ".jpg") {
						if strings.Contains(imgLink, "https") {
							article.ImageLink = imgLink
							return false
						}
					}
					return true
				})
				log.Println(article.Date + " " + forum + "版-" + "標題: (" + article.LikeCountString + ")" + article.Title)
				db.InsertArticle(article.Title, article.LikeCount, article.Link, article.Date, article.ImageLink, article.LikeCountString, article.Board)
			}
		})
		crawlerCount = crawlerCount + 1
	}

}

func passR18(reqURL string) (req *http.Request) {

	// 設定 header 以及 滿18歲cookie
	req, _ = http.NewRequest("GET", reqURL, nil)

	over18Cookie := http.Cookie{
		Name:       "over18",
		Value:      "1",
		Domain:     "www.ptt.cc",
		Path:       "/",
		RawExpires: "Session",
		MaxAge:     0,
		HttpOnly:   false,
	}

	req.AddCookie(&over18Cookie)

	return req

}
