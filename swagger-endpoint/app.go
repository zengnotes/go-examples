package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/xmlpath.v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

var stubbedResponseJson string = `
{
  "name": "GOOGL",
  "last": 1002.20,
  "time": "12:34",
  "date": "10/31/2014"
}
`

type Response struct {
	Name string  `json:name"Name"`
	Last float64 `json:name"Last"`
	Time string  `json:name"Time"`
	Date string  `json:name"Date"`
}

var soapStart string = `
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.webserviceX.NET/">
<soapenv:Header/><soapenv:Body><web:GetQuote><web:symbol>`

var soapEnd string = `</web:symbol></web:GetQuote></soapenv:Body></soapenv:Envelope>`

func getQuote(r *http.Request) (string, error) {

	//Get the symbol as the last part of the request URI
	parts := strings.Split(r.RequestURI, "/")
	symbol := parts[len(parts)-1]
	fmt.Printf("requesting quote for %s\n", symbol)

	payload := fmt.Sprintf("%s%s%s", soapStart, symbol, soapEnd)

	client := &http.Client{}
	quoteReq, err := http.NewRequest("POST", "http://www.webservicex.net/stockquote.asmx", strings.NewReader(payload))
	if err != nil {
		return "", err
	}

	quoteReq.Header.Add("Content-Type", "text/xml")

	resp, err := client.Do(quoteReq)
	if err != nil {
		return "", err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	resp.Body.Close()
	return string(respData), nil

}

func getQuoteResponse(soapResponse string) (string, error) {
	compiledPath := xmlpath.MustCompile("/Envelope/Body/GetQuoteResponse/GetQuoteResult")
	root, err := xmlpath.Parse(strings.NewReader(soapResponse))
	if err != nil {
		return "", err
	}

	value, ok := compiledPath.String(root)
	if !ok {
		return "", errors.New("Unable to extract GetQuoteResult")
	}

	quoteResult := string(value)

	return quoteResult, nil
}

func getQuotePart(quoteResult string, partName string) (string, error) {
	compiledPath := xmlpath.MustCompile("/StockQuotes/Stock/" + partName)
	root, err := xmlpath.Parse(strings.NewReader(quoteResult))
	if err != nil {
		log.Fatal(err)
	}

	value, ok := compiledPath.String(root)
	if !ok {
		return "", errors.New("Unable to extract " + partName)
	}

	return string(value), nil
}

func formResponseJson(data string) ([]byte, error) {

	//Pull out the quote result part of the soap envelope
	quoteResult, err := getQuoteResponse(data)
	if err != nil {
		return nil, err
	}

	//From the quote result, pull out the name
	name, err := getQuotePart(quoteResult, "Name")
	if err != nil {
		return nil, nil
	}

	//Grab the last value
	last, err := getQuotePart(quoteResult, "Last")
	if err != nil {
		return nil, nil
	}

	lastFloat, err := strconv.ParseFloat(last, 64)
	if err != nil {
		return nil, nil
	}

	//Time
	time, err := getQuotePart(quoteResult, "Time")
	if err != nil {
		return nil, nil
	}

	//Date
	date, err := getQuotePart(quoteResult, "Date")
	if err != nil {
		return nil, nil
	}

	response := Response{
		Name: name,
		Last: lastFloat,
		Time: time,
		Date: date,
	}

	return json.Marshal(response)
}

func handleQuoteCalls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")

	quoteData, err := getQuote(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	responseJson, err := formResponseJson(quoteData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(responseJson))
}

func corsWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(rec.Body.Bytes())
	})
}

func main() {
	//Original static content
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/stuff/", http.StripPrefix("/stuff/", fs))

	//Add some dynamc content
	http.HandleFunc("/quote/", handleQuoteCalls)

	//Add the swagger spec
	ss := http.FileServer(http.Dir("dist"))
	http.Handle("/apispec/", http.StripPrefix("/apispec/", corsWrapper(ss)))

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}