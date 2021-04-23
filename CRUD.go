package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: loadCA("cert/ca.crt")},
	},
}

const baseURL = "https://localhost:5000/api/v1/courses"

//course structure that is stored in the database.
type course struct {
	CourseID  string
	Title     string
	Lecturer  string
	ClassSize int
}

//regular expression pattern for user input.
var (
	regexCourseID      = regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`)
	regexTitleLecturer = regexp.MustCompile(`^[\w\d\s]{3,30}$`) //same regex format can be used for course title and lecturer
	regexClassSize     = regexp.MustCompile(`^[0-9]{1,4}$`)
)

//addCourse take in all four required inputs  from user. Empty input is not allowed.
func addCourse() {

	var courseID, title, lecturer, classsize string
	for courseID == "" {
		fmt.Println("Please provide the course ID.")
		fmt.Scanln(&courseID)
	}
	courseID = Policy.Sanitize(strings.TrimSpace(courseID)) // input validation and sanitization
	if !regexCourseID.MatchString(courseID) {
		log.Error("Incorrect input format for Course ID detected. --addCourse")
		return
	}

	for title == "" {
		fmt.Println("Please provide the course title.")
		input := bufio.NewReader(os.Stdin)
		title, _ = input.ReadString('\n')
		title = strings.TrimRight(title, "\n")
	}
	title = Policy.Sanitize(strings.TrimSpace(title)) // input validation and sanitization
	if !regexTitleLecturer.MatchString(title) {
		log.Error("Incorrect input format for Course Title detected. --addCourse")
		return
	}

	for lecturer == "" {
		fmt.Println("Please provide the lecturer name of the course.")
		input := bufio.NewReader(os.Stdin)
		lecturer, _ = input.ReadString('\n')
		lecturer = strings.TrimRight(lecturer, "\n")
	}
	lecturer = Policy.Sanitize(strings.TrimSpace(lecturer)) // input validation and sanitization
	if !regexTitleLecturer.MatchString(lecturer) {
		log.Error("Incorrect input format for Course Lecturer detected. --addCourse")
		return
	}

	var classsizeInt int
	for classsize == "" {
		fmt.Println("Please provide expected class size.")
		fmt.Scanln(&classsize)
	}
	classsize = Policy.Sanitize(strings.TrimSpace(classsize)) // input validation and sanitization
	if !regexClassSize.MatchString(classsize) {
		log.Error("Incorrect input format for Class Size detected. --addCourse")
		return
	}
	classsizeInt, _ = strconv.Atoi(classsize)
	if classsizeInt <= 0 {
		log.Error("Class Size must be greater than zero. --addCourse")
		return
	}

	jsonData := course{courseID, title, lecturer, classsizeInt}
	fmt.Println(jsonData)
	jsonValue, _ := json.Marshal(jsonData)
	//response, err := http.Post(baseURL+"/"+courseID+"?key="+key, "application/json", bytes.NewBuffer(jsonValue))
	response, err := client.Post(baseURL+"/"+courseID+"?key="+key, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		log.Error("The HTTP request failed with error %s\n", err, "--addCourse")
	} else {
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
	}
}

//getCourse take in course ID as input. If no course ID is specified, it will be show the user the whole list of course.
//The user input was obtained upfront in console menu function.
func getCourse(courseID string) []byte {

	var url string
	if courseID != "" {
		url = baseURL + "/" + courseID + "?key=" + key
	} else {
		url = baseURL + "?key=" + key
	}

	response, err := client.Get(url)
	//response, err := http.Get(url)
	var data []byte
	if err != nil {
		log.Error("The HTTP request failed with error: ", err, "  --getCourse")
	} else {
		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)
		if courseID != "" {
			fmt.Println(response.StatusCode)
			fmt.Println(string(data))
		}
	}
	return data
}

//updateCourse check with the user which field required to be updated,
//if the user do not wish to update a particular field, he/she can press enter to skip that field.
func updateCourse() {

	var courseID, titleUpdated, lecturerUpdated, classsizeUpdated string
	var jsonData course

	fmt.Println("Please provide the course ID.")
	fmt.Scanln(&courseID)
	courseID = Policy.Sanitize(strings.TrimSpace(courseID)) // input validation and sanitization
	if !regexCourseID.MatchString(courseID) {
		log.Error("Incorrect input format for Course ID detected. --updateCourse")
		return
	}

	//Retrieve course details once obtained courseID
	data := getCourse(courseID)
	json.Unmarshal(data, &jsonData)

	fmt.Println("Please provide the course title. Please enter if there is no change.")
	input := bufio.NewReader(os.Stdin)
	titleUpdated, _ = input.ReadString('\n')
	titleUpdated = strings.TrimRight(titleUpdated, "\n")
	if titleUpdated != "" {
		titleUpdated = Policy.Sanitize(strings.TrimSpace(titleUpdated)) // input validation and sanitization
		if !regexTitleLecturer.MatchString(titleUpdated) {
			log.Error("Incorrect input format for Course Title detected. --updateCourse")
			return
		}
		jsonData.Title = titleUpdated
	}

	fmt.Println("Please provide the lecturer name of the course.Please enter if there is no change.")
	input = bufio.NewReader(os.Stdin)
	lecturerUpdated, _ = input.ReadString('\n')
	lecturerUpdated = strings.TrimRight(lecturerUpdated, "\n")
	if lecturerUpdated != "" {
		lecturerUpdated = Policy.Sanitize(strings.TrimSpace(lecturerUpdated)) // input validation and sanitization
		if !regexTitleLecturer.MatchString(lecturerUpdated) {
			log.Error("Incorrect input format for Course Lecturer detected. --updateCourse")
			return
		}
		jsonData.Lecturer = lecturerUpdated
	}

	fmt.Println("Please provide expected class size.Please enter if there is no change.")
	fmt.Scanln(&classsizeUpdated)
	if classsizeUpdated != "" {
		classsizeUpdated = Policy.Sanitize(strings.TrimSpace(classsizeUpdated)) // input validation and sanitization
		if !regexClassSize.MatchString(classsizeUpdated) {
			log.Error("Incorrect input format for Class Size detected. --updateCourse")
			return
		}
		classsizeUpdatedInt, _ := strconv.Atoi(classsizeUpdated)
		if classsizeUpdatedInt <= 0 {
			log.Error("Class Size must be greater than zero. --updateCourse")
			return
		}
		jsonData.ClassSize = classsizeUpdatedInt
	}

	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest(http.MethodPut, baseURL+"/"+courseID+"?key="+key, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("The HTTP request failed with error: ", err, "--updateCourse")
	}
	request.Header.Set("Content-Type", "application/json")

	//client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		log.Error("The HTTP request failed with error: ", err, "--updateCourse")
	} else {
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
	}
}

//deleteCourse perform course deletion through the input of course ID by user.
func deleteCourse() {

	var courseID string
	fmt.Println("Please provide the course ID you wish to delete.")
	fmt.Scanln(&courseID)
	courseID = Policy.Sanitize(strings.TrimSpace(courseID)) // input validation and sanitization
	if !regexCourseID.MatchString(courseID) {
		log.Error("Incorrect input format for Course ID detected. --deleteCourse")
		return
	}

	request, err := http.NewRequest(http.MethodDelete, baseURL+"/"+courseID+"?key="+key, nil)
	if err != nil {
		log.Error("The HTTP request failed with error: ", err, "--deleteCourse")
	}
	//client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error("The HTTP request failed with error: ", err, "--deleteCourse")
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
		response.Body.Close()
	}
}
