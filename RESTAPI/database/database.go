package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type Course struct {
	CourseID  string
	Title     string
	Lecturer  string
	ClassSize int
}

var ctx = context.Background()

func CourseExist(db *sql.DB, CourseID string) (int, error) {
	query := fmt.Sprintln("SELECT EXISTS(SELECT * FROM Course WHERE CourseID=?)")
	var exist int
	err := db.QueryRowContext(ctx, query, CourseID).Scan(&exist)
	if err != nil {
		return 0, err
	}
	return exist, err
}

func DeleteRecord(db *sql.DB, CourseID string) error {
	query := fmt.Sprintln("DELETE FROM Course WHERE CourseID=?")
	_, err := db.QueryContext(ctx, query, CourseID)
	return err
}

func EditRecord(db *sql.DB, CourseID string, Title string, Lecturer string, ClassSize int) error {
	query := fmt.Sprintln("UPDATE Course SET Title=?, Lecturer=?, ClassSize=? WHERE CourseID=?")
	_, err := db.QueryContext(ctx, query, Title, Lecturer, ClassSize, CourseID)
	return err
}

func InsertRecord(db *sql.DB, CourseID string, Title string, Lecturer string, ClassSize int) error {
	query := fmt.Sprintln("INSERT INTO Course VALUES (?, ?, ?, ?)")
	_, err := db.QueryContext(ctx, query, CourseID, Title, Lecturer, ClassSize)
	return err
}

func GetRecord(db *sql.DB, CourseID string) (Course, error) {
	query := fmt.Sprintln("SELECT * FROM Course WHERE CourseID=?")
	var course Course
	err := db.QueryRowContext(ctx, query, CourseID).Scan(&course.CourseID, &course.Title, &course.Lecturer, &course.ClassSize)
	return course, err
}

// map this type to the record in the table
func GetAllRecords(db *sql.DB) []Course {
	allCourses := []Course{}
	results, err := db.Query("Select * FROM my_db_goMicroservice1.Course")
	if err != nil {
		log.Panic(err.Error())
	}
	for results.Next() { //.Next go through every single record
		// map this type to the record in the table
		var course Course
		err = results.Scan(&course.CourseID, &course.Title, &course.Lecturer, &course.ClassSize)
		if err != nil {
			log.Panic(err.Error())
		}
		allCourses = append(allCourses, course)
	}
	return allCourses
}
