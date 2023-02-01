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
	"time"
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
	dbNet := flag.String("dbNet", "", "Net for the Database.")
	ipAddr := flag.String("ipAddr", "", "IP Address for the Database.")
	dbName := flag.String("dbName", "", "Name of the Database")
	keyFileName := flag.String("keyFile", "", "A file containing a list of keywords")
	outputFileName := flag.String("outputFile", "", "A file containing a list of keywords")
	lastTaskTimeFileName := flag.String("lastTaskTimeFileName", "", "A file containing the time when the last checked task was created")
	flag.Parse()
	// Get DB connection
	connectDB(*dbUser, *dbPass, *dbNet, *ipAddr, *dbName)
	// Get keywords from user provided file
	keywords := getKeywordsFromFile(*keyFileName)
	// Declare a slice to hold all error tasks
	var extractedErrorTasks []string
	var updatedTaskTime string
	var errorTasks []string
	var numOfNewErrors int
	// Get the created time of the last checked task
	givenLastTaskTime := readLatestTaskTimeFile(*lastTaskTimeFileName)
	foundNewErrors := false
	// if no keywords were provided call the function with an empty string
	if len(keywords) == 0 {
		errorTasks, updatedTaskTime = getErrorTasks("", givenLastTaskTime)
		if len(errorTasks) != 0 {
			foundNewErrors = true
			numOfNewErrors = len(errorTasks)
		}
		extractedErrorTasks = append(extractedErrorTasks, errorTasks...)
		// if keywords were provided call the function with each keyword
	} else {
		for _, k := range keywords {
			returnedTaskTime := ""
			errorTasks, returnedTaskTime = getErrorTasks(k, givenLastTaskTime)
			// If no error task is found in an iteration because the given keyword does not exist, don't default back to the old updatedTaskTime
			if returnedTaskTime != "None" {
				updatedTaskTime = returnedTaskTime
			}
			// I think the problem is that if no error task is found it returns the updatedTask which was given which is a problem
			extractedErrorTasks = append(extractedErrorTasks, errorTasks...)
		}
		if len(extractedErrorTasks) != 0 {
			foundNewErrors = true
			numOfNewErrors = len(extractedErrorTasks)
		}
	}
	// Write the error tasks to a file
	file := getFile(*outputFileName)
	defer file.Close()
	if foundNewErrors {
		writeTasksToFile(file, extractedErrorTasks)
		updateLatestTaskTimeFile(*lastTaskTimeFileName, updatedTaskTime)
		fmt.Printf("Found %d new errors and updated the file %s with the time %s\n", numOfNewErrors, *lastTaskTimeFileName, updatedTaskTime)
	} else {
		fmt.Println("No new errors since ", givenLastTaskTime)
	}
}

func connectDB(user string, pass string, net string, ipAddr string, dbName string) *sql.DB {
	cfg := mysql.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  net,
		Addr:                 ipAddr,
		DBName:               dbName,
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

func getErrorTasks(keyword string, givenLastTaskTime string) ([]string, string) {
	// An errorTasks slice to hold data from returned rows.
	var errorTasks []string
	// A variable to hold a "random" dummy time for the first task, which is ensured to be before the first task in the database
	lastTaskTime := givenLastTaskTime
	// If no keywords are provided query every error task which has a std_error output
	if keyword == "" {
		// TODO Only get tasks which are newer than the last task time
		rows, err := db.Query("SELECT taskUUID, createdTime, stdError FROM Tasks WHERE stdError IS NOT NULL AND stdError != '' AND createdTime > ? ORDER BY createdTime DESC", lastTaskTime)
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var task ErrorTask
			if err := rows.Scan(&task.TaskUUID, &task.CreatedAt, &task.StdError); err != nil {
				log.Fatal(err)
			}
			if isTaskNew(task.CreatedAt, lastTaskTime) {
				lastTaskTime = task.CreatedAt
			}
			//updatedLastTaskTime = getLatestTaskTime(task.CreatedAt, lastTaskTime)
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
		rows, err := db.Query("SELECT taskUUID, createdTime, stdError FROM Tasks WHERE stdError IS NOT NULL AND stdError != '' AND stdError COLLATE utf8_general_ci LIKE ? AND createdTime > ? ORDER BY createdTime DESC", "%"+keyword+"%", lastTaskTime)
		if !rows.Next() {
			lastTaskTime = "None"
		}
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var task ErrorTask
			if err := rows.Scan(&task.TaskUUID, &task.CreatedAt, &task.StdError); err != nil {
				log.Fatal(err)
			}
			if isTaskNew(task.CreatedAt, lastTaskTime) {
				lastTaskTime = task.CreatedAt
			}
			//updatedLastTaskTime = getLatestTaskTime(task.CreatedAt, lastTaskTime)
			taskJson, err := json.Marshal(task)
			if err != nil {
				log.Fatal(err)
			}
			// Add the task to the errorTasks slice.
			errorTasks = append(errorTasks, string(taskJson))
		}
		defer rows.Close()
	}
	return errorTasks, lastTaskTime
}

func isTaskNew(createdTask string, givenTask string) bool {
	createdTaskTime := convertStringToTime(createdTask)
	givenTaskTime := convertStringToTime(givenTask)
	// if the given createdTime is after the latestTaskTime return the createdTime as the new latestTaskTime
	if createdTaskTime.After(givenTaskTime) {
		return true
	} else {
		return false
	}
}

func convertStringToTime(createdAt string) time.Time {
	// Convert string to time.Time
	layout := "2006-01-02 15:04:05.000000"
	t, err := time.Parse(layout, createdAt)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func getFile(fileName string) *os.File {
	// Create file if it doesn't exist, otherwise open it.
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		return file
	} else {
		// Open file for writing only, append and not overwrite existing, readable and writable by everyone.
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		return file
	}
}

func writeTasksToFile(file *os.File, errorTasks []string) {
	// Write tasks to the file.
	for _, task := range errorTasks {
		_, err := file.WriteString(task + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func updateLatestTaskTimeFile(lastTaskTimeFileName string, lastTaskTime string) {
	// This writes a new file which tells us the created time of the latest Task we got from the database.
	// This allows us to only get tasks which have been created after this time.
	file, err := os.Create(lastTaskTimeFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Write tasks to the file.
	_, err = file.WriteString(lastTaskTime)
	if err != nil {
		log.Fatal(err)
	}
}

func readLatestTaskTimeFile(lastTaskTimeFileName string) string {
	// Open the file
	file, err := os.Open(lastTaskTimeFileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lastTaskTime string
	for scanner.Scan() {
		// Print the keyword
		lastTaskTime = scanner.Text()
	}
	return lastTaskTime
}

// TODO: If a keyword is added which is not in the error log we run in this weird problem that the lastTaskTime is not updated.
