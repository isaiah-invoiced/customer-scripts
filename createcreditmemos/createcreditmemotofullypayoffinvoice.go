package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"os"
	"strings"
)

//This program will create credit memos on Invoiced; one per row
// rows must be of the form:
// [0]: {invoice_number}

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will create credit memos to cancel out invoices from the excel file")

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "" {
		fmt.Println("Enter P for Production, S for Sandbox: ")
		*environment, _ = reader.ReadString('\n')
		*environment = strings.TrimSpace(*environment)
	}

	if *environment == "P" {
		sandBoxEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" {
		fmt.Println("Using Sandbox for the environment")
	} else {
		fmt.Println("Unrecognized value ", *environment, ", enter P or S only")
		return
	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	invoiceNumberIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No invoice numbers in excel sheet to create credit memos")
	}

	rows = rows[1:len(rows)]

	for _, row := range rows {

		invoiceNumberParsed := strings.TrimSpace(row[invoiceNumberIndex])

		invoice, err := conn.NewInvoice().ListInvoiceByNumber(invoiceNumberParsed)

		if err != nil {
			fmt.Println("Error retrieving value of credit note;" +
				"invoice #"+ invoiceNumberParsed +" does not exist")
			continue
		}

		if invoice == nil {
			fmt.Println("Error retrieving value of credit note;" +
				"invoice #"+ invoiceNumberParsed +" does not exist")
			continue
		}

		// create simplified items to pass into credit note
		// if we don't do this, request will fail
		items := make([]invdendpoint.LineItem, len(invoice.Items))
		for k, v := range invoice.Items {
			items[k].Name = v.Name
			items[k].Quantity = v.Quantity
			items[k].UnitCost = v.UnitCost
		}

		creditNote := conn.NewCreditNote()
		creditNote.Customer = invoice.Customer
		creditNote.Invoice = invoice.Id
		creditNote.Items = items

		_, err = creditNote.Create(creditNote)

		if err != nil {
			fmt.Println("Error creating credit note for invoice " + invoiceNumberParsed +
				" - error: " + err.Error())
			continue
		}

		fmt.Println("Successfully created & issued credit note for invoice " + invoiceNumberParsed)
	}



}
