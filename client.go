// Client for latex2gmd by Calvin Liu
// Processes a Latex input file, then sends a tokenized version of the input to the server

// Note: If program cannot find the import packages for some reason, copy the following into ~/.bashrc then run source ~/.bashrc
// 	 export GOROOT=/usr/lib/go
// 	 export GOPATH=$HOME/go
// 	 export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

// based on https://www.rabbitmq.com/tutorials/tutorial-six-go.html

package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// failOnError - checks that every action was successful
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err) // errors out of the program
	}
}

// Rabbit-mq Section

// sendRPC - sends a message to the server
func sendRPC(request []token) (res map[string]int, err error) {

	// start connection, close automatically when function ends
	conn, err := amqp.Dial("amqp://localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// start channel, close automatically when function ends
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare queue and register consumer
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrID := uuid.New().String()
	var body []byte

	body, _ = json.Marshal(request)

	// send list of tokens to server
	err = ch.Publish(
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			ReplyTo:       q.Name,
			Body:          body,
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrID == d.CorrelationId {
			json.Unmarshal(d.Body, &res)
			failOnError(err, "Failed to convert body")
			break
		}
	}

	return
}

// Tokenize Latex Section

// numWorkers - value represents the number of goroutine workers active when the program is run
// increase the number for more concurrent threads, and decrease for less
const numWorkers = 8

// lineWorker - structure for assigning jobs to the workers
type lineWorker struct {
	Order   int
	RawData string
}

// Token Section

// token - struct for storing the parsed data
type token struct {
	Order          int
	Data           string
	ToggleMathMode bool
}

// updateToken - updates the values inside the token
func (tk *token) updateToken(order int, data string, toggleMathMode bool) {
	tk.Order = order
	tk.Data = data
	tk.ToggleMathMode = toggleMathMode
}

// parseLine - parses the line for keywords and returns a token that can be processed later
// mostly tested Github Flavored Markdown with https://jbt.github.io/markdown-editor/
// used several modified examples from https://www.javatpoint.com/latex-fractions
func tokenizeLine(order int, line string) token {
	var res token

	// the order of which substring is checked first can be rearranged for minor
	// speed up optimizations based on the most common options found in Latex
	if len(line) > 0 {
		if strings.Contains(line, "\\begin{equation") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{equation") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\begin{align") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{align") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\begin{flalign") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{flalign") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\begin{multline") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{multline") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\begin{math") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{math") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\begin{center") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\end{center") {
			res.updateToken(order, "", true)
			return res
		}
		if strings.Contains(line, "\\documentclass") {
			return res
		}
		if strings.Contains(line, "\\usepackage") {
			return res
		}
		if strings.Contains(line, "\\begin{document") {
			return res
		}
		if strings.Contains(line, "\\end{document") {
			return res
		}
		if strings.Contains(line, "\\maketitle") {
			return res
		}
		if strings.Contains(line, "\\label") {
			return res
		}
		if strings.Contains(line, "\\begin{abstract") {
			res.updateToken(order, "## Abstract", false)
			return res
		}
		if strings.Contains(line, "\\end{abstract") {
			return res
		}
		if strings.Contains(line, "\\title") {
			res.updateToken(order, strings.Trim("# "+line[7:len(line)-1], " "), false)
			return res
		}
		if strings.Contains(line, "\\author") {
			res.updateToken(order, strings.Trim("By: "+line[8:len(line)-1], " "), false)
			return res
		}
		if strings.Contains(line, "\\section") {
			res.updateToken(order, strings.Trim("## "+line[9:len(line)-1], " "), false)
			return res
		}
		if strings.Contains(line, "\\subsection") {
			res.updateToken(order, strings.Trim("### "+line[12:len(line)-1], " "), false)
			return res
		}
		if strings.Contains(line, "%") {
			pos := strings.Index(line, "%")
			res.updateToken(order, strings.Trim(line[:pos], " "), false)
			return res
		}
	}

	// default case if there are no keywords found in the line
	res.updateToken(order, strings.Trim(line, " "), false)

	return res
}

// worker - function to process each line and tokenize it
func worker(wg *sync.WaitGroup, jobs <-chan lineWorker, results chan<- interface{}) {
	defer wg.Done()

	for j := range jobs {
		res := tokenizeLine(j.Order, j.RawData)
		results <- res
	}
}

// tokenizeLatex - will tokenize a latex file and return a list of tokens
// adapted from https://brandur.org/go-worker-pool
// adapted from https://stackoverflow.com/questions/27217428/reading-a-file-concurrently-in-golang
func tokenizeLatex(inputFile string) []token {

	// opens the input file for read access
	file, err := os.Open(inputFile)
	failOnError(err, "Failed to open Latex file")

	// create buffered channels for assigning jobs and receiving results
	jobs := make(chan lineWorker)
	results := make(chan interface{})

	// create a wait group to synchronize all jobs
	var wg sync.WaitGroup

	// start up some workers that are ready to parse lines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs, results)
	}

	// go over a file line by line and queue up a ton of work
	go func() {
		counter := 0
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			jobs <- lineWorker{counter, scanner.Text()}
			counter++
		}
		close(jobs)
	}()

	// collect all the results, and close the result channel afterwards
	go func() {
		wg.Wait()
		close(results)
	}()

	// generate a list of tokens to send over RPC
	var tokenList []token

	for r := range results {
		res := r.(token)

		tokenList = append(tokenList, res)
	}

	// synchronize all the concurrent work and sort them into the order that
	// they appeared in the input file
	// adapted from https://stackoverflow.com/questions/28999735/what-is-the-shortest-way-to-simply-sort-an-array-of-structs-by-arbitrary-field
	sort.Slice(tokenList, func(i, j int) bool {
		return tokenList[i].Order < tokenList[j].Order
	})

	return tokenList
}

// Main Function Section

// reads input file name from the bash script
func main() {
	argumentList := os.Args

	if len(argumentList) < 2 {
		// checks that the user actually provides an input file
		// redundant check since the bash file also requires an input file
		// will only be seen in the case that this is not called from ./latex2gmd.sh
		println("Please enter in file name to convert")

	} else {
		println("Processing Latex file")
		tokenList := tokenizeLatex(argumentList[1])

		res, err := sendRPC(tokenList)
		failOnError(err, "Failed to handle RPC request")

		// outputs the date and time of message
		log.Printf("%d", res["length"])
	}
}
