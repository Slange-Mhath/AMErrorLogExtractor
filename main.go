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
	"strings"
)

var db *sql.DB

type ErrorTask struct {
	TaskUUID  string `field:"task_uuid"`
	CreatedAt string `field:"created_at"`
	StdError  string `field:"std_error"`
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:62001",
		DBName:               "MCP",
		AllowNativePasswords: true,
	}

	fileName := "errorTasks.json"

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// Get the Tasks from the DB which threw an error.
	errorTasks, err := getErrorTasks()

	// Write the error tasks to a file
	err = writeToFile(fileName, errorTasks)
	if err != nil {
		return
	}

}

func getErrorTasks() ([]string, error) {
	// An errorTasks slice to hold data from returned rows.
	var errorTasks []string
	// Query the database.
	rows, err := db.Query("SELECT taskUUID, createdTime, stdError FROM Tasks WHERE stdError IS NOT NULL AND stdError != ''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keywords := getKeywordsFromFile()

	// Iterate over the rows, sending one message per row.
	for rows.Next() {
		var task ErrorTask
		if err := rows.Scan(&task.TaskUUID, &task.CreatedAt, &task.StdError); err != nil {
			return nil, err
		}
		// Check if keywords were given in the file and look for them accordingly in the error message.
		if len(keywords) > 0 {
			for _, keyword := range keywords {
				if checkIfKeywordInError(keyword, task) {
					// Convert the task to JSON.
					taskJson, err := json.Marshal(task)
					if err != nil {
						return nil, err
					}
					// Add the task to the errorTasks slice.
					errorTasks = append(errorTasks, string(taskJson))
				}
			}
			// If no keywords were given in the file, just add all the tasks to the errorTasks slice.
		} else {
			taskJson, err := json.Marshal(task)
			if err != nil {
				return nil, err
			}
			// Add the task to the errorTasks slice.
			errorTasks = append(errorTasks, string(taskJson))
		}

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return errorTasks, nil
}

func writeToFile(fileName string, errorTasks []string) error {
	// Create the file.
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	// Write tasks to the file.
	for _, task := range errorTasks {
		_, err := file.WriteString(task)
		if err != nil {
			return err
		}
	}
	return nil
}

func getKeywordsFromFile() []string {
	//keyword := flag.String("keyword", "", "Keyword to check for in the error message.")
	//flag.Parse()
	//fmt.Println(*keyword)

	// Create a new flag set
	flags := flag.NewFlagSet("keywords", flag.ExitOnError)

	// Add a flag to specify the file
	fileName := flags.String("file", "", "A file containing a list of keywords")

	// Parse the flags
	flags.Parse(os.Args[1:])

	// Open the file
	file, err := os.Open(*fileName)
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

func checkIfKeywordInError(keyword string, task ErrorTask) bool {
	// If keyword is empty its safe to assume that it wont be in the error message.
	if keyword == "" {
		return false
		// If the keyword is in the error message return true, before the check both strings get lowered to be case insensitive.
	} else if strings.Contains(strings.ToLower(task.StdError), strings.ToLower(keyword)) {
		return true
	} else {
		return false
	}
}

// TODO: Or we have a main function which takes an error keyword and makes an SQL Query to the Task Table where
// STDError contains the keyword. And then add those return structs to a list which we then write to a file.
// Maybe thats a better approach as it wont cause a for loop in a for loop
