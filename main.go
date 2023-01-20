package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var db *sql.DB

// Define the ErrorTask we want to add to a slice and put into a list
type ErrorTask struct {
	TaskUUID  string `field:"task_uuid"`
	CreatedAt string `field:"created_at"`
	StdError  string `field:"std_error"`
}

func main() {
	// Get user input
	dbPass := flag.String("dbPass", "", "Password for the Database.")
	dbUser := flag.String("dbUser", "", "User for the Database.")
	keyFileName := flag.String("keyFile", "", "A file containing a list of keywords")
	outputFileName := flag.String("outputFile", "", "A file containing a list of keywords")
	flag.Parse()
	// Get DB connection
	connectDB(*dbUser, *dbPass)
	// Get keywords from user provided file
	keywords := getKeywordsFromFile(*keyFileName)
	// Declare a slice to hold all error tasks
	var extractedErrorTasks []string
	// if no keywords were provided call the function with an empty string
	if len(keywords) == 0 {
		errorTasks := getErrorTasks("")
		extractedErrorTasks = append(extractedErrorTasks, errorTasks...)
		// if keywords were provided call the function with each keyword
	} else {
		for _, k := range keywords {
			errorTasks := getErrorTasks(k)
			extractedErrorTasks = append(extractedErrorTasks, errorTasks...)
		}
	}
	// Write the error tasks to a file
	writeToFile(*outputFileName, extractedErrorTasks)
}

func connectDB(user string, pass string) *sql.DB {
	cfg := mysql.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  "tcp",
		Addr:                 "127.0.0.1:62001",
		DBName:               "MCP",
		AllowNativePasswords: true,
	}
	// Get database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func getKeywordsFromFile(fileName string) []string {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var keywords []string
	for scanner.Scan() {
		// Print the keyword
		keywords = append(keywords, scanner.Text())
	}
	return keywords

}

func getErrorTasks(keyword string) []string {
	// An errorTasks slice to hold data from returned rows.
	var errorTasks []string
	// If no keywords are provided query every error task which has a std_error output
	if keyword == "" {
		rows, err := db.Query("SELECT taskUUID, createdTime, stdError FROM Tasks WHERE stdError IS NOT NULL AND stdError != ''")
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var task ErrorTask
			if err := rows.Scan(&task.TaskUUID, &task.CreatedAt, &task.StdError); err != nil {
				log.Fatal(err)
			}
			taskJson, err := json.Marshal(task)
			if err != nil {
				log.Fatal(err)
			}
			// Add the task to the errorTasks slice.
			errorTasks = append(errorTasks, string(taskJson))
		}
		rows.Close()
		// If keywords are provided query every error task which has a std_error output and contains the keyword
	} else {
		rows, err := db.Query("SELECT taskUUID, createdTime, stdError FROM Tasks WHERE stdError IS NOT NULL AND stdError != '' AND stdError COLLATE utf8_general_ci LIKE ?", "%"+keyword+"%")

		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var task ErrorTask
			if err := rows.Scan(&task.TaskUUID, &task.CreatedAt, &task.StdError); err != nil {
				log.Fatal(err)
			}
			taskJson, err := json.Marshal(task)
			if err != nil {
				log.Fatal(err)
			}
			// Add the task to the errorTasks slice.
			errorTasks = append(errorTasks, string(taskJson))
		}
		defer rows.Close()
	}

	return errorTasks
}

func writeToFile(fileName string, errorTasks []string) {
	// Create the file.
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Write tasks to the file.
	for _, task := range errorTasks {
		_, err := file.WriteString(task)
		if err != nil {
			log.Fatal(err)
		}
	}
}
