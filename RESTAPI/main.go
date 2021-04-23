package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"

	database "goMicroService1Assignment/RESTAPI/database"
)

var (
	db                                             *sql.DB
	dbPort, dbHost, dbUsername, dbPassword, dbName string
	APIKey                                         string
	Port                                           string
	//Unique policy creation for the life of the program.
	Policy = bluemonday.UGCPolicy()
)

//regular expression pattern for user input.
var (
	regexCourseID      = regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`)
	regexTitleLecturer = regexp.MustCompile(`^[\w\d\s]{3,30}$`) //same regex format can be used for course title and lecturer
)

//validKey function verify the incoming API key in the request is valid.
func validKey(w http.ResponseWriter, r *http.Request) bool {
	v := r.URL.Query()
	if key, ok := v["key"]; ok {
		if key[0] == APIKey {
			return true
		} else { //invalid key
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("401 - Invalid key"))
			log.Error("Fail attempt in providing API key: 401 - Invalid key")
			return false
		}
	} else { //key is not provided  //code is modified to specify the exact error
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Please supply access key"))
		log.Error("Fail attempt in providing API key: 401 - Please supply access key")
		return false
	}
}

//home lead to the homepage of the API
func home(w http.ResponseWriter, r *http.Request) {

	if !validKey(w, r) {
		return
	}

	fmt.Fprintf(w, "Welcome to the REST API!")
}

//allcourses allow the function to response to the console application with all courses information.
func allcourses(w http.ResponseWriter, r *http.Request) {

	if !validKey(w, r) {
		return
	}

	allCourses := database.GetAllRecords(db)

	// returns all the courses in JSON
	json.NewEncoder(w).Encode(&allCourses)

}

//course function will perform the necessary CRUD operation based on the HTTP method in the request.
func course(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic("Panic occured at console menu and recovered.", err)
		}
	}()

	if !validKey(w, r) {
		return
	}

	if r.Method == "GET" {

		params["courseid"] = Policy.Sanitize(params["courseid"]) // input validation and sanitization
		if !regexCourseID.MatchString(params["courseid"]) {
			log.Error("Incorrect format for Course ID detected. --getRecord")
			return
		}

		course, err := database.GetRecord(db, params["courseid"])
		//fmt.Println(course)
		if err == nil {
			json.NewEncoder(w).Encode(&course)
		} else if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
			log.Warning("Fail attempt to get record: 404 - No course found")
		} else {
			log.Panic(err.Error())
		}
	}

	if r.Method == "DELETE" {

		params["courseid"] = Policy.Sanitize(params["courseid"]) // input validation and sanitization
		if !regexCourseID.MatchString(params["courseid"]) {
			log.Error("Incorrect format for Course ID detected. --deleteRecord")
			return
		}

		exist, err := database.CourseExist(db, params["courseid"])
		if err != nil {
			log.Panic(err.Error())
		} else if exist != 0 {
			err := database.DeleteRecord(db, params["courseid"])
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Error in deleteing course!"))
				log.Error("Fail attempt to delete record: 422 - Error in deleteing course!")
			} else {
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("202 - Course deleted: " + params["courseid"]))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
			log.Warning("Fail attempt to delete record: 404 - No course found")
		}
	}

	if r.Header.Get("Content-type") == "application/json" { //only incoming POST and PUT is expected to be in JSON format
		// POST is for creating new course
		if r.Method == "POST" {
			var newCourse database.Course
			// read the string sent to the service
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// convert JSON to object
				json.Unmarshal(reqBody, &newCourse)

				// input validation and sanitization before sent to insert into a sql query
				params["courseid"] = Policy.Sanitize(params["courseid"])
				if !regexCourseID.MatchString(params["courseid"]) {
					log.Error("Incorrect format for Course ID detected. --insertRecord")
					return
				}

				if newCourse.Title == "" || newCourse.Lecturer == "" || newCourse.ClassSize <= 0 {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Information supplied not complete. Please supply course information in JSON format."))
					log.Warning("Fail attempt to insert record: 422 - Please supply course information in JSON format")
					return
				}
				// check if course exists; add only if course does not exist
				exist, err := database.CourseExist(db, params["courseid"])
				if err != nil {
					log.Panic(err.Error())
				} else if exist == 0 {
					// input validation and sanitization before sent to insert into a sql query
					newCourse.Title = Policy.Sanitize(strings.TrimSpace(newCourse.Title))
					if !regexTitleLecturer.MatchString(newCourse.Title) {
						log.Error("Incorrect format for Course Title detected. --insertRecord")
						return
					}
					// input validation and sanitization before sent to insert into a sql query
					newCourse.Lecturer = Policy.Sanitize(strings.TrimSpace(newCourse.Lecturer))
					if !regexTitleLecturer.MatchString(newCourse.Lecturer) {
						log.Error("Incorrect format for Course Lecturer detected. --insertRecord")
						return
					}

					database.InsertRecord(db, params["courseid"], newCourse.Title, newCourse.Lecturer, newCourse.ClassSize)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Course added: " + params["courseid"]))
				} else {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("409 - Duplicate course ID"))
					log.Warning("Fail attempt to insert record: 409 - Duplicate course ID")
				}
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply course information in JSON format"))
				log.Warning("Fail attempt to insert record: 422 - Please supply course information in JSON format")
			}
		}

		// PUT is for creating or updating existing course
		if r.Method == "PUT" {
			var newCourse database.Course
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				json.Unmarshal(reqBody, &newCourse)

				// input validation and sanitization before sent to insert into a sql query
				params["courseid"] = Policy.Sanitize(params["courseid"])
				if !regexCourseID.MatchString(params["courseid"]) {
					log.Error("Incorrect format for Course ID detected. --insertRecord")
					return
				}

				if newCourse.Title == "" || newCourse.Lecturer == "" || newCourse.ClassSize <= 0 {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte("422 - Please supply course information in JSON format"))
					log.Warning("Fail attempt to insert record: 422 - Please supply course information in JSON format")
					return
				}

				// check if course exists; add only if course does not exist
				exist, err := database.CourseExist(db, params["courseid"])
				if err != nil {
					log.Panic(err.Error())
				} else if exist != 0 {
					// input validation and sanitization before sent to insert into a sql query
					newCourse.Title = Policy.Sanitize(strings.TrimSpace(newCourse.Title))
					if !regexTitleLecturer.MatchString(newCourse.Title) {
						log.Error("Incorrect format for Course Title detected. --editRecord")
						return
					}
					// input validation and sanitization before sent to insert into a sql query
					newCourse.Lecturer = Policy.Sanitize(strings.TrimSpace(newCourse.Lecturer))
					if !regexTitleLecturer.MatchString(newCourse.Lecturer) {
						log.Error("Incorrect format for Course Lecturer detected. --editRecord")
						return
					}

					database.EditRecord(db, params["courseid"], newCourse.Title, newCourse.Lecturer, newCourse.ClassSize)
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - Course updated: " + params["courseid"]))
				} else if exist == 0 {
					database.InsertRecord(db, params["courseid"], newCourse.Title, newCourse.Lecturer, newCourse.ClassSize)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Course added: " + params["courseid"]))
				}
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply course information in JSON format"))
				log.Warning("Fail attempt to insert record: 422 - Please supply course information in JSON format")
			}
		}
	}

}

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

//getting all the information from .env file
func init() {

	APIKey = goDotEnvVariable("APIKEY")
	dbUsername = goDotEnvVariable("dbUsername")
	dbPassword = goDotEnvVariable("dbPassword")
	dbHost = goDotEnvVariable("dbHost")
	dbPort = goDotEnvVariable("dbPort")
	dbName = goDotEnvVariable("dbName")
	Port = goDotEnvVariable("port")
}

func main() {

	// Use mysql as driverName and a valid DSN as dataSourceName:
	var err error
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, err = sql.Open("mysql", dataSourceName)

	// handle error
	if err != nil {
		log.Panic(err.Error())
	}
	defer db.Close()

	//database.GetAllRecords(db)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/", home).Schemes("https")
	router.HandleFunc("/api/v1/courses", allcourses).Schemes("https")
	router.HandleFunc("/api/v1/courses/{courseid}", course).Methods("GET", "PUT", "POST", "DELETE").Schemes("https")

	fmt.Println("Listening at port 5000")
	//log.Fatal(http.ListenAndServe(":5000", router))

	//connect to the port according to the assignment requirement
	err1 := http.ListenAndServeTLS(":"+Port, "cert/server.crt", "cert/server.key", router)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

// use godot package to load/read the .env file and return the value of the key
func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)

}
