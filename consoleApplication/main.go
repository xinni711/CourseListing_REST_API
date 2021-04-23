package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
)

var (
	key string
	//Unique policy creation for the life of the program.
	Policy = bluemonday.UGCPolicy()
)

func init() {

	// Create the log file if doesn't exist. And append to it if it already exists.
	file, err := os.OpenFile("log/logfile.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	} else {
		log.SetOutput(io.MultiWriter(file, os.Stdout)) //default logger will be writing to file and os.Stdout
	}
	Formatter := new(log.TextFormatter)
	log.SetLevel(log.WarnLevel) //only log the warning severity level or higher
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true

}

//loadCA will load root CA that will verify the server certificate
func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()

	if ca, e := ioutil.ReadFile(caFile); e != nil {
		log.Fatal("ReadFile: ", e)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

func main() {

	key = goDotEnvVariable("APIKEY") //obtain API key from the environment variable file.
	consoleMenu()

}

//consoleMenu will display main page of CRUD operation selection for course listing.
func consoleMenu() {

	var choice int

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic("Panic occured at console menu and recovered.", err)
		}
	}()

	for choice != 6 {

		fmt.Println("\n=================================================")
		fmt.Println("University Course Listing Page (Lecturer Access)")
		fmt.Println("=================================================")
		fmt.Println("1. Add a new course")
		fmt.Println("2. Browse selected course")
		fmt.Println("3. Browse all course")
		fmt.Println("4. Edit existing course")
		fmt.Println("5. Delete existing course")
		fmt.Println("6. Exit the course listing page")
		fmt.Println("Select your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			addCourse()
		case 2:
			var courseID string
			for courseID == "" {
				fmt.Println("Please provide the course ID you wish to browse.")
				fmt.Scanln(&courseID)
			}
			getCourse(courseID)
			courseID = Policy.Sanitize(strings.TrimSpace(courseID))
			if !regexCourseID.MatchString(courseID) {
				log.Warning("Incorrect input format for Course ID detected. --getCourse")
				return
			}
		case 3:
			var jsonData []interface{}
			data := getCourse("")
			json.Unmarshal(data, &jsonData)
			fmt.Println("\nBelow is the list of available course.")
			for i, v := range jsonData {
				fmt.Printf("%d. Course ID: %s, Title: %s, Lecturer: %s, Class Size: %v \n",
					i+1, v.(map[string]interface{})["CourseID"], v.(map[string]interface{})["Title"],
					v.(map[string]interface{})["Lecturer"], v.(map[string]interface{})["ClassSize"])
			}
		case 4:
			updateCourse()
		case 5:
			deleteCourse()
		case 6:
			fmt.Println("Exiting the booking system")
		default:
			fmt.Println("Please select 1 to 6.")
		}

	}

}

//goDotEnvVariable use godot package to load/read the .env file and return the value of the key
func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv(key)

}
