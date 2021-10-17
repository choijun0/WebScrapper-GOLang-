package main

import (
  "fmt"
  "github.com/PuerkitoBio/goquery"
  "net/http"
  "log"
  "strconv"
  "strings"
  "os"
  "encoding/csv"
  "errors"
  "github.com/labstack/echo/v4"
)


type extractedInfo struct {
  id string
  title string
  company string 
  location string
  salary string
  summary string 
} 

//Scrapper Package

func Scrape(term string) {
  baseURL := "https://kr.indeed.com/jobs?q=" + term;
  pages := getPages(baseURL);
  var jobsInfo []extractedInfo
  jobsInfoConChannel := make(chan []extractedInfo)

  //Request
  for i:=0; i<pages; i++ {
    go getPage(i, baseURL, jobsInfoConChannel)
  }
  //Recieve
  for i:=0; i<pages; i++{
    jobsInfo=append(jobsInfo, <- jobsInfoConChannel...);
  }
  fmt.Println("Done. extracted "+strconv.Itoa(pages));

  writeJobs(jobsInfo, baseURL);
}


//form the url of each page extracted by getPages func and by using them exrtact extract more info
func getPage(page int, baseURL string, c chan<- []extractedInfo) {
  pageURL := "";
  var jobInfoCon []extractedInfo;
  jobInfoChannel := make(chan extractedInfo);

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
  cards := doc.Find(".tapItem")
  //Request
  cards.Each(func(i int, card *goquery.Selection){
    go extractJonInfo(card, jobInfoChannel)
  })
  //Receive
  for i:=0; i<cards.Length(); i++ {
    jobInfoCon = append(jobInfoCon, <- jobInfoChannel);
  }

  c <- jobInfoCon;
}


func extractJonInfo(card *goquery.Selection, c chan<- extractedInfo) {
    id, _ := card.Attr("id");

    title, _ := card.Find(".jobTitle>span").Attr("title");
    title = CleaningString(title);

    //companyName Element is exist in both span>a and span only
    companyElement := card.Find(".companyName")  
    if companyElement.Length() == 0 {
      companyElement = card.Find(".companyName>a")
    }
    company := CleaningString(companyElement.Text());
    
		location := CleaningString(card.Find(".companyLocation").Text())

    summary := CleaningString(card.Find(".job-snippet").Text())
    
    salary := CleaningString(card.Find(".salary-snippet-container>span").Text())
    c <- extractedInfo{
      id: id, 
      title: title,
      company: company, 
      location: location,
      summary: summary, 
      salary: salary}
}

func CleaningString(str string) string{
  //#.1 Trimspace를 이용하여 문자열의 양옆 공백제거
  //#.2 Fields 를 이용해 text만을 요소로하는 문자배열생성
  //#.3 Join을 이용해 " "을 사이에두고 문자배열을 문자열로 합침
  return strings.Join(strings.Fields(strings.TrimSpace(str))," ");
}

//Get html file of baseURL(By goquery) and Exract page count
func getPages(baseURL string) int{
  res, err := http.Get(baseURL);
  defer res.Body.Close()
  checkErr(err);
  checkStatus(res);
  doc, err := goquery.NewDocumentFromReader(res.Body)
  checkErr(err);
  //CardsInPage
  crdpCount := doc.Find(".tapItem").Length()

  //AllCards
  pageCountElement := doc.Find("div#searchCountPages");
  if pageCountElement.Length() == 0 {
    checkErr(errors.New("can't find page count"));
  }
  slice := strings.Split(pageCountElement.Text(), " ");

  lastElement := slice[len(slice)-1]; //페이지수 + 건
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

func writeJobs(jobsInfo []extractedInfo, baseURL string) {
  file, err := os.Create("jobs.csv");
  finish := make(chan bool);
  checkErr(err);

  w:= csv.NewWriter(file);
  defer w.Flush()

  headers := []string{"URL", "TITLE", "COMPANY", "LOCATION", "SALARY", "SUMMARY"}
  w.Write(headers);
  for _, job := range jobsInfo {
    go writeToMain(w, job, finish, baseURL);
  }

  //block flush when writing isn't done.
  for i:=0; i<len(jobsInfo); i++ {
    <- finish 
  }
}

func writeToMain(w *csv.Writer, job extractedInfo, c chan bool, baseURL string) {
  jobSlice := []string{baseURL + "&vjk=" + job.id[4 : len(job.id) -1], job.title,job.company, job.location, job.salary, job.summary}
  err := w.Write(jobSlice)
  checkErr(err)
  c <- true;
}

//##########################################################################

var fileName string = "jobs.csv"

func main() {
  e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1123"))
}

func handleHome(c echo.Context) error{
  return c.File("home.html")
}

func handleScrape(c echo.Context) error {
  //defer os.Remove(fileName)
	term := strings.ToLower(CleaningString(c.FormValue("term")))
  Scrape(term);
	return c.Attachment(fileName, fileName) 
}