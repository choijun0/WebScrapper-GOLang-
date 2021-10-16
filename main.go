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
)

var baseURL = "https://kr.indeed.com/jobs?q=javascript";

type extractedInfo struct {
  id string
  title string
  company string 
  location string
  salary string
  summary string 
} 

func main() {
  //pages := getPages();
  var jobsInfo []extractedInfo
  for i:=0; i<5; i++ {
    //exrtact func(http.Get) is usynchronized func so it and for func can't be come together
    jobsInfo=append(jobsInfo, getPage(i)...); //the way spread array in go
  }
  writeJobs(jobsInfo);
}

func writeJobs(jobsInfo []extractedInfo) {
  file, err := os.Create("jobs.csv");
  checkErr(err);

  w:= csv.NewWriter(file);
  defer w.Flush()

  headers := []string{"URL", "TITLE", "COMPANY", "LOCATION", "SALARY", "SUMMARY"}
  w.Write(headers);

  for _, job := range jobsInfo {

    jobSlice := []string{"https://kr.indeed.com/jobs?q=javascript&vjk=" + job.id[4 : len(job.id) -1], job.title,job.company, job.location, job.salary, job.summary}
    err := w.Write(jobSlice)
    checkErr(err)
  }
}


//form the url of each page extracted by getPages func and by using them exrtact extract more info
func getPage(page int) []extractedInfo {
  pageURL := "";
  var jobInfoCon []extractedInfo;
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
  cards.Each(func(i int, card *goquery.Selection){
    jobInfo := extractJonInfo(card);
    jobInfoCon = append(jobInfoCon, jobInfo);
  })

  return jobInfoCon;
}


func extractJonInfo(card *goquery.Selection) extractedInfo{
    id, _ := card.Attr("id");

    title, _ := card.Find(".jobTitle>span").Attr("title");
    title = cleaningString(title);
    
    company_location := card.Find(".company_location>pre");
    company := cleaningString(company_location.Find(".companyName>a").Text())
		location := cleaningString(company_location.Find(".companyLocation").Text())

    //can't extract!!!
    summary := cleaningString(card.Find(".snippet").Text())
    
    salary := cleaningString(card.Find(".salary-snippet-container>span").Text())

    return extractedInfo{
      id: id, 
      title: title,
      company: company, 
      location: location,
      summary: summary, 
      salary: salary}
}

func cleaningString(str string) string{
  //#.1 Trimspace를 이용하여 문자열의 양옆 공백제거
  //#.2 Fields 를 이용해 text만을 요소로하는 문자배열생성
  //#.3 Join을 이용해 " "을 사이에두고 문자배열을 문자열로 합침
  return strings.Join(strings.Fields(strings.TrimSpace(str))," ");
}



//Get html file of baseURL(By goquery) and Exract page count
func getPages() int{
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