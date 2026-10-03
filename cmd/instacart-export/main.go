package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	instacart "github.com/beezyfbaby/go-instacart-export"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "instacart-export:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("instacart-export", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", filepath.Join("data", "instacart_deliveries_"+time.Now().Format("2006-01-02_15-04-05.000000000")+".csv"), "CSV destination (must not already exist)")
	queryHash := flags.String("query-hash", instacart.PersonalOrderHistoryHash, "PersonalOrderHistory persisted-query SHA-256 hash")
	timeout := flags.Duration("timeout", 5*time.Minute, "maximum time for entire export")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	client := instacart.Client{SessionToken: os.Getenv("INSTACART_SESSION_TOKEN"), HostSessionToken: os.Getenv("INSTACART_HOST_SESSION_TOKEN")}
	fmt.Fprintln(stderr, "Fetching orders from the undocumented consumer endpoint...")
	orders, err := instacart.FetchDeliveryHistory(ctx, client, *queryHash)
	if err != nil {
		return err
	}
	if err := writeToCSV(*output, extractDeliveryData(orders)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Exported %d deliveries to %s\n", len(orders), *output)
	return err
}

func extractOrdersData(orders []*instacart.Order) [][]string {
	data := [][]string{{"id", "status", "total", "createdAt", "retailers", "numItems"}}
	for _, o := range orders {
		var retailers []string
		numItems := 0
		for _, d := range o.Deliveries {
			retailers = append(retailers, d.Retailer)
			numItems += len(d.Items)
		}
		row := []string{o.ID, o.Status, o.Total, o.CreatedAt.Format(time.DateOnly), strings.Join(retailers, "|"), strconv.Itoa(numItems)}
		for i := range row {
			row[i] = spreadsheetSafe(row[i])
		}
		data = append(data, row)
	}
	return data
}

// Prevent spreadsheet applications from interpreting untrusted text as formulas.
func spreadsheetSafe(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") || (len(trimmed) > 0 && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
		return "'" + value
	}
	return value
}

func writeToCSV(path string, data [][]string) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	writer := csv.NewWriter(file)
	if err = writer.WriteAll(data); err != nil {
		return err
	}
	return file.Sync()
}

// Each row represents a delivery. Do not merge or sum multi-order totals.
func extractDeliveryData(deliveries []instacart.HistoryDelivery) [][]string {
	data := [][]string{{"deliveryId", "orderId", "status", "deliveryTotal", "createdAt", "retailer", "numItemLines", "isMulti", "serviceType"}}
	for _, d := range deliveries {
		row := []string{d.ID, d.OrderID, d.Status, d.Total, d.CreatedAt.Format(time.RFC3339Nano), d.Retailer, strconv.Itoa(len(d.Items)), strconv.FormatBool(d.IsMulti), d.ServiceType}
		for i := range row {
			row[i] = spreadsheetSafe(row[i])
		}
		data = append(data, row)
	}
	return data
}
