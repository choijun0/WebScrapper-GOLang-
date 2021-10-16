package main

import (
  "fmt"
  "github.com/PuerkitoBio/goquery"
  "net/http"
  "log"
  "strconv"
  "strings"
)

var baseURL = "https://kr.indeed.com/jobs?q=javascript";

func main() {
  pages := getPagesItemCount();

  for i:=0; i<pages; i++ {
    //exrtact func(http.Get) is usynchronized func so it and for func can't be come together
    getPage(i)
  }
}

//form the url of each page extracted by getPages func and by using them exrtact extract more info
func getPage(page int) {
  pageURL := "";
  if page != 0{
    pageURL = baseURL + "&start=" + strconv.Itoa(page * 10);
  } else {
    pageURL = baseURL;
  }
  fmt.Println("Requesting ", pageURL);
  res, err := http.Get(pageURL);
  defer res.Body.Close()
  checkErr(err)
  checkStatus(res);
  doc, _ := goquery.NewDocumentFromReader(res.Body);
  cards := doc.Find(".tapItem");
  fmt.Println(cards.Length())
  cards.Each(func(i int, card *goquery.Selection){
    id, _ := card.Attr("id");
    infoCon := card.Find(".resultContent");

    title, _ := infoCon.Find(".jobTitle>span").Attr("title");
    
    company_location := infoCon.Find(".company_location>pre");
    company := company_location.Find(".companyName>a").Text()
		location := company_location.Find(".companyLocation").Text()
    
    fmt.Println(id, title, company, location);
  })
}

//Get html file of baseURL(By goquery) and Exract page count
func getPagesItemCount() int{
  res, err := http.Get(baseURL);
  defer res.Body.Close()
  checkErr(err);
  checkStatus(res);
  doc, err := goquery.NewDocumentFromReader(res.Body)
  checkErr(err);
  //CardsInPage
  crdpCount := doc.Find(".tapItem").Length()

  //AllCards
  slice := strings.Split(doc.Find("div#searchCountPages").Text(), " ");
  lastElement := slice[len(slice)-1]; //페이지수 + 건
  fmt.Println(lastElement);
  cardCountStr := lastElement[0 : len(lastElement) - 3]; //"건" 추출
  acrdCount, _ := strconv.Atoi(cardCountStr);

  //Pages
  pages := int(acrdCount / crdpCount) + ((acrdCount % crdpCount) % 2)
  return pages;
}

func checkErr(err error){
  if err != nil {
    log.Fatalln(err)
  }
}

func checkStatus(res *http.Response) {
  if res.StatusCode != 200 {
    log.Fatalln("Request Failed statusCode :", res.StatusCode);
  }
}