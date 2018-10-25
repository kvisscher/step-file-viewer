package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type Value struct {
	AttributeID string `xml:"AttributeID,attr"`
	Text        string `xml:",innerxml"`
}

type ValueGroup struct {
	AttributeID string  `xml:"AttributeID,attr"`
	Values      []Value `xml:"Value"`
}

type ProductCrossReference struct {
	ID     string  `xml:"ProductID,attr"`
	Type   string  `xml:"Type,attr"`
	Values []Value `xml:"MetaData>Value"`
}

type Product struct {
	ID             string `xml:"ID,attr"`
	ParentID       string
	Name           string
	Values         []Value                 `xml:"Values>Value"`
	ValueGroup     []ValueGroup            `xml:"Values>ValueGroup"`
	CrossReference []ProductCrossReference `xml:"ProductCrossReference"`
	Children       []Product               `xml:"Product"`
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("missing one or more arguments")
	}

	searchField := strings.TrimSpace(os.Args[1])
	searchValue := strings.TrimSpace(os.Args[2])

	f := path.Clean(os.Args[3])

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current working directory %v", err)
	}

	var products []Product

	cachedFile := path.Join(dir, path.Base(f))
	if _, err := os.Stat(cachedFile); !os.IsNotExist(err) {
		// Don't have to parse XML
		log.Printf("reading cached file %s", cachedFile)

		if b, err := ioutil.ReadFile(cachedFile); err == nil {
			if err = json.Unmarshal(b, &products); err != nil {
				log.Fatalf("failed to parse cache %v", err)
			}
		} else {
			log.Fatalf("failed to read cache %v", err)
		}

		log.Println("read cache")
	} else {
		log.Printf("parsing products in file %s..", f)
		products = parseProducts(f)
		log.Printf("parsed %d products", len(products))

		log.Println("saving cache..")
		f1, _ := json.MarshalIndent(products, "", "  ")
		ioutil.WriteFile(cachedFile, []byte(f1), 0666)
		log.Println("saved cache")
	}

	log.Printf("going to search for products that matches %s = %s", searchField, searchValue)

	var matchedProducts []Product
	for _, p := range products {
		if searchInValues(searchField, searchValue, &p) {
			log.Printf("found a match %s %s", p.ID, p.Name)

			matchedProducts = append(matchedProducts, p)
		}
	}

	log.Printf("going to output %d files..", len(matchedProducts))

	for _, p := range matchedProducts {
		fileName := strings.Replace(fmt.Sprintf("step-%s-%s.json", p.ID, p.Name), " ", "_", -1)
		contents, _ := json.MarshalIndent(p, "", "  ")

		ioutil.WriteFile(path.Join(dir, fileName), []byte(contents), 0444)
	}
}

func parseProducts(fullPathToFile string) []Product {
	f, err := os.Open(fullPathToFile)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	fi, err := f.Stat()

	if err != nil {
		log.Fatalf("could not stat file: %v", err)
	}

	decoder := xml.NewDecoder(bufio.NewReaderSize(f, 1<<16))
	var products []Product

	size := fi.Size()
	lastProgress := 0.0

	var lastTime float64
	start := time.Now()

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		offset := decoder.InputOffset()
		progress := (float64(offset) / float64(size)) * 100

		if progress != lastProgress {
			lastProgress = progress
		}

		timePassed := time.Now().Sub(start).Seconds()

		if timePassed-lastTime > 1 {
			lastTime = timePassed
			log.Printf("%.0f percent", progress)
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Product" {
				var p Product

				decoder.DecodeElement(&p, &se)

				if p.ID != "" {
					products = append(products, p)
				}
			}
		}
	}

	for _, p := range products {
		products = recurseAddChildren(p, products)
	}

	return products
}

func recurseAddChildren(product Product, products []Product) []Product {
	for i, p := range product.Children {
		p.ParentID = product.ID

		product.Children[i] = p

		products = recurseAddChildren(p, products)

		if p.ID != "" {
			products = append(products, p)
		}
	}

	return products
}

func searchInValues(property, value string, product *Product) bool {
	for _, v := range product.Values {
		if strings.EqualFold(v.AttributeID, property) && v.Text == value {
			return true
		}
	}

	for _, g := range product.ValueGroup {
		if strings.EqualFold(g.AttributeID, property) {
			for _, v := range g.Values {
				if v.Text == value {
					return true
				}
			}
		}
	}

	for _, p := range product.Children {
		return searchInValues(property, value, &p)
	}

	return false
}
