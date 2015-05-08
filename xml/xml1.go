package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

var xml1Doc = `
<foodoc>
	<type>Thing1</type>
	<things>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
		<thing>xxx</thing>
		<thing>yyy</thing>
	</things>
</foodoc>
`

type Result struct {
	XMLName xml.Name `xml:"foodoc"`
	Type    string   `xml:"type"`
	Things  []string `xml:"things>thing"`
}

func xml1(verbose bool) {

	v := new(Result)

	err := xml.Unmarshal([]byte(xml1Doc), &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	if verbose {
		fmt.Printf("Type: %s\n", v.Type)
		fmt.Printf("Things: %v\n", v.Things)
	}

}

func streamParseXml1(verbose bool) {
	reader := strings.NewReader(xml1Doc)
	decoder := xml.NewDecoder(reader)
	v := new(Result)
	var typeData = false
	var thingData = false

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch t.(type) {
		default:
			fmt.Println("Not sure what we're dealing with here")
		case xml.StartElement:

			if verbose {
				fmt.Println("StartElement")
				fmt.Printf("\tElement: %s\n", t.(xml.StartElement).Name)
			}
			elementName := t.(xml.StartElement).Name.Local
			switch elementName {
			case "type":
				typeData = true
			case "thing":
				thingData = true
			}

		case xml.EndElement:
			if verbose {
				fmt.Println("EndElement")
			}

			elementName := t.(xml.EndElement).Name.Local
			switch elementName {
			case "type":
				typeData = false
			case "thing":
				thingData = false
			}

		case xml.CharData:
			if verbose {
				fmt.Println("CharData")
				fmt.Println(string(t.(xml.CharData)))
			}

			if typeData {
				v.Type = string(t.(xml.CharData))
			} else if thingData {
				v.Things = append(v.Things, string(t.(xml.CharData)))

			}
		}

	}
	if verbose {
		fmt.Printf("%v\n", v)
	}
}
