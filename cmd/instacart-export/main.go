package main

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	instacart "github.com/beezyfbaby/go-instacart-export"
)

func main() {
	sessionToken := os.Getenv("INSTACART_SESSION_TOKEN")

	if sessionToken == "" {
		log.Fatal("Session token missing. Please provide the INSTACART_SESSION_TOKEN environment variable")
	}

	client := instacart.Client{
		SessionToken: sessionToken,
	}

	log.Print("Fetching orders...")
	orders := instacart.FetchOrders(client)
	data := extractOrdersData(orders)
	writeToCSV(data)

	log.Print("Done!")
}

func extractOrdersData(orders []*instacart.Order) [][]string {
	log.Print("Processing orders")
	data := [][]string{{
		"id",
		"status",
		"total",
		"createdAt",
		"retailers",
		"numItems",
	}}
	for _, o := range orders {

		var retailers []string
		numItems := 0

		for _, d := range o.Deliveries {
			retailers = append(retailers, d.Retailer)
			numItems += len(d.Items)
		}

		order := []string{
			o.ID,
			o.Status,
			o.Total,
			o.CreatedAt.Format("2006-01-02"),
			strings.Join(retailers, "|"),
			strconv.Itoa(numItems),
		}
		data = append(data, order)
	}

	return data
}

func writeToCSV(data [][]string) {
	log.Print("Writing orders to a CSV")
	path := "data"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatal("Unable to create directory", err)
		}
	}
	now := time.Now()
	file, err := os.Create("data/instacart_orders_" + now.Format("01-02-2006_03-04-05") + ".csv")
	if err != nil {
		log.Fatal("Unable to create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			log.Fatal("Error writing data", err)
		}
	}
}
