package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func authMiddleWare(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized."})
		c.Abort()
		return
	}
	c.Next()
}

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func connectDB() {
	var err error

	url := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Println("Connect to database error", err)
		return
	}
}

func createTable() {
	createTb := `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err := db.Exec(createTb)

	if err != nil {
		fmt.Println("Unable to create table")
		return
	}

	fmt.Println("Create Table Success")
}

func postCustomer(c *gin.Context) {
	cust := Customer{}

	err := c.ShouldBindJSON(&cust)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "JSON parsing on insertion error!!! " + err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customer (name,email,status) values ($1,$2,$3) RETURNING id", cust.Name, cust.Email, cust.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Unable to insert record"})
		return
	}

	cust.ID = id
	c.JSON(http.StatusCreated, cust)
}

func getCustByID(c *gin.Context) {
	id := c.Param("id")
	cust := Customer{}

	stmt, err := db.Prepare("SELECT id, name, email,status FROM customer where id=$1;")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on prepare query for get record by id " + err.Error()})
		return
	}

	row := stmt.QueryRow(id)

	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on query record by id " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func getAllCust(c *gin.Context) {
	cust := Customer{}
	result := []Customer{}

	stmt, err := db.Prepare("SELECT id, name, email,status FROM customer;")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on prepare query for get all record " + err.Error()})
		return
	}

	rows, err := stmt.Query()

	for rows.Next() {
		err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Unable to get all records" + err.Error()})
			return
		}

		result = append(result, cust)
	}

	c.JSON(http.StatusOK, result)
}

func putCustByID(c *gin.Context) {
	id := c.Param("id")
	cust := Customer{}

	err := c.ShouldBindJSON(&cust) //JSON.Unmarshall(data,&t)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "JSON parsing on update error!!! " + err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE customer SET name=$2, email=$3, status=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on prepare query for update record by id " + err.Error()})
		return
	}

	if _, err := stmt.Exec(id, cust.Name, cust.Email, cust.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on update record by id " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func deleteCustByID(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("DELETE FROM customer WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on prepare query for delete record by id " + err.Error()})
		return
	}

	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Error on delete record by id " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func main() {
	//connect to db on elephantSQL
	connectDB()
	defer db.Close()

	//Create customer table
	createTable()

	r := gin.Default()

	//for authorized check
	r.Use(authMiddleWare)

	// Rest API - 5 end points
	r.POST("/customers", postCustomer)
	r.GET("/customers/:id", getCustByID)
	r.GET("/customers", getAllCust)
	r.PUT("/customers/:id", putCustByID)
	r.DELETE("/customers/:id", deleteCustByID)

	r.Run(":2019")
}
