package main

import (
  "fmt"
  "github.com/PuerkitoBio/goquery"
  "net/http"
  "log"
  "strconv"
)

var baseURL = "https://kr.indeed.com/jobs?q=javascript";

func main() {
  itemCount := getPagesItemCount();

  for i:=0; ; i++ {
    if getPage(i, itemCount) != false {
      break;
    }
  }
}

//form the url of each page extracted by getPages func and by using them exrtact extract more info
func getPage(page int, itemCount int) bool {
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
    //check is LastPage
  if cards.Length() != itemCount {
    return true;
  }
  return false;
}

//Get html file of baseURL(By goquery) and Exract page count
func getPagesItemCount() int{
  res, err := http.Get(baseURL);
  defer res.Body.Close()
  checkErr(err);
  checkStatus(res);
  doc, err := goquery.NewDocumentFromReader(res.Body)
  checkErr(err);
  //get num of selections
  itemCount := doc.Find(".tapItem").Length();

  return itemCount;
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