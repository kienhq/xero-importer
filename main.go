package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var env = make(map[string]string)

type Account struct {
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	Type        string `json:"Type"`
	Description string `json:"Description,omitempty"`
}

type Processor struct {
	accounts []*Account
	workers  []*Worker
	counter  int32
	doneCh   chan struct{}
}

type Worker struct {
	accessToken string
	tenantId    string
	client      *http.Client
	accountCh   chan *Account
	processor   *Processor
}

func NewWorker(accessToken string, tenantId string, client *http.Client, processor *Processor) *Worker {
	return &Worker{
		accessToken: accessToken,
		tenantId:    tenantId,
		client:      client,
		accountCh:   make(chan *Account, 1),
		processor:   processor,
	}
}

func (w *Worker) Start() {
	for {
		select {
		case account := <-w.accountCh:
			w.ProcessAccount(account, 0)
		}
	}
}

func NewProcessor(accessToken, tenantId string, numberOfWorker int, accounts []*Account) *Processor {
	p := &Processor{
		accounts: accounts,
		doneCh:   make(chan struct{}),
		workers:  make([]*Worker, numberOfWorker),
	}
	for i := 0; i < numberOfWorker; i++ {
		p.workers[i] = NewWorker(accessToken, tenantId, http.DefaultClient, p)
	}
	return p
}

func (p *Processor) Start() {
	for _, worker := range p.workers {
		go worker.Start()
	}
	go p.processAccounts(p.accounts)
	for {
		select {
		case <-p.doneCh:
			atomic.AddInt32(&p.counter, 1)
			if atomic.LoadInt32(&p.counter) >= int32(len(p.accounts)) {
				return
			}
		}
	}
}

func (p *Processor) processAccounts(accounts []*Account) {
	for _, account := range accounts {
		p.sendToWorker(account)
	}
}

func (p *Processor) sendToWorker(account *Account) {
	for {
		for _, worker := range p.workers {
			select {
			case worker.accountCh <- account:
				return
			default:
				continue
			}
		}
		time.Sleep(time.Second)
	}
}

func (w *Worker) ProcessAccount(account *Account, count int) {
	if count < 10 {
		err := w.uploadAccount(account)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			w.ProcessAccount(account, count+1)
		} else {
			log.Println(fmt.Sprintf("finish processing account %s (uploaded: %d)", account.Code, atomic.LoadInt32(&w.processor.counter)))
		}
	} else {
		log.Println(fmt.Sprintf("cannot process account %s (uploaded: %d)", account.Code, atomic.LoadInt32(&w.processor.counter)))
	}
	w.processor.doneCh <- struct{}{}
}

func (w *Worker) uploadAccount(account *Account) error {
	url := "https://api.xero.com/api.xro/2.0/Accounts"
	accountData := map[string]string{
		"Code":        account.Code,
		"Name":        account.Name,
		"Type":        account.Type,
		"Description": account.Description,
	}
	accountJSON, err := json.Marshal(accountData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(accountJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+w.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Xero-Tenant-Id", w.tenantId)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		if strings.Contains(bodyString, "Please enter a unique") {
			return nil
		}
		return fmt.Errorf("failed to upload account: %v %v", resp.Status, bodyString)
	}

	return nil
}

func readCSV(filename string) ([]*Account, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var accounts []*Account
	for _, record := range records[1:] { // Skip header row
		account := &Account{
			Code:        record[0],
			Name:        record[1],
			Type:        strings.ToUpper(record[2]), // Convert account type to uppercase
			Description: record[4],
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func generateCOA(numberOfFile int, path string, initNumber, size int) {
	n := 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	types := []string{
		"Inventory",
		"Expense",
		"Fixed",
		"Revenue",
		"Current",
		"Currliab",
		"Depreciatn",
		"DirectCosts",
		"Equity",
		"Liability",
		"NonCurrent",
		"Otherincome",
		"Overheads",
		"Prepayment",
		"Sales",
		"Termliab",
	}

	descriptions := []string{
		"Value of tracked items for resale.",
		"Standard-Rated Purchases (8%),An expenditure that has been paid for in advance.",
		"An expenditure that has been paid for in advance.",
		"Unrealised currency gains on outstanding items",
		"Gains or losses made due to currency exchange rate changes",
		"A percentage of total earnings paid to the government.",
		"Outstanding invoices the company has issued out to the client but has not yet received in cash at" +
			" balance date.",
	}
	counter := 0
	var files []string
	for n < numberOfFile {
		filename := filepath.Join(path, fmt.Sprintf("coa%d.csv", n))
		f, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		_, err = f.WriteString("*Code,*Name,*Type,*Tax Code,Description,Dashboard,Expense Claims,Enable Payments,Balance\n")
		if err != nil {
			panic(err)
		}
		for i := 1; i <= size; i++ {
			id := initNumber + counter + 1
			name := fmt.Sprintf("Test%d", id)
			tax := "No Tax (0%)"
			_, err = f.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,No,No,Yes,\n", id, name, types[r.Intn(len(types)-1)], tax, descriptions[r.Intn(len(descriptions)-1)]))
			if err != nil {
				panic(err)
			}
			counter++
		}
		files = append(files, filename)
		n++
		f.Close()
	}
	if len(files) > 0 {
		env["FILES"] = strings.Join(files[:], ",")
		err := godotenv.Write(env, "./.env")
		if err != nil {
			panic(err)
		}
	}
}

func uploadAccounts(accessToken, tenantId string, files []string) {
	accounts := make([]*Account, 0)
	for _, file := range files {
		acc, err := readCSV(strings.TrimSpace(file))
		if err != nil {
			log.Println(err)
		} else {
			accounts = append(accounts, acc...)
		}
	}

	processor := NewProcessor(accessToken, tenantId, 2, accounts)
	processor.Start()
}

func main() {
	var err error
	env, err = godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("COA_PATH") != "" {
		env["COA_PATH"] = os.Getenv("COA_PATH")
	}

	accessToken := env["ACCESS_TOKEN"]
	tenantId := env["TENANT_ID"]
	files := strings.Split(env["FILES"], ",")

	// env for generate accounts files
	numberOfGeneratedFiles, err := strconv.Atoi(env["NUM_GENERATED_FILES"])
	if err != nil {
		panic(err)
	}
	numberOfGeneratedCOA, err := strconv.Atoi(env["NUM_GENERATED_COA"])
	if err != nil {
		panic(err)
	}
	initCOANumber, err := strconv.Atoi(env["INIT_COA_NUMBER"])
	if err != nil {
		panic(err)
	}
	coaPath := env["COA_PATH"]

	switch os.Getenv("ACTION") {
	case "upload_accounts":
		uploadAccounts(accessToken, tenantId, files)
	case "generate_accounts":
		generateCOA(numberOfGeneratedFiles, coaPath, initCOANumber, numberOfGeneratedCOA)
	default:
		panic("invalid action")
	}
}
