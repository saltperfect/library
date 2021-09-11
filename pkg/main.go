package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

// id, name, isbn
var (
	dbhost     = "localhost"
	dbport     = "53"
	dbuser     = "go"
	dbpassword = "go"
	databasename   = "library"
)

type library struct {
	host, user, password, dbname string
	port                         int
}

type Book struct {
	Id, Name, Isbn string
}

const (
	API_PATH = "/api/v1/books"
)

func main() {

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = dbhost
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = dbport
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = dbuser
	}
	password := os.Getenv("DB_PASS")
	if password == "" {
		password = dbpassword
	}
	dbname:= os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = databasename
	}

	intport, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	lib := library{host: host, user: user, password: password, dbname: dbname, port: intport}

	// fmt.Printf("%+v\n",lib)
	db := lib.createConnection()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc(API_PATH, lib.getBooks).Methods("GET")
	r.HandleFunc(API_PATH, lib.postBooks).Methods("POST")
	log.Print("starting server \n")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("while serving application %s", err.Error())
	}
}
func (l library) postBooks(w http.ResponseWriter, r *http.Request) {
	log.Println("post call recieved")
	db := l.createConnection()
	defer l.closeConnection(db)
	var books []Book
	dec := json.NewDecoder(r.Body)
	dec.Decode(&books)

	fmt.Println(books)

	for _, book := range books {
		_, err := db.Exec("insert into books(id, name, isbn) values ($1, $2, $3)", book.Id, book.Name, book.Isbn)
		if err != nil {
			continue
		}
	}

}

func (l library) getBooks(w http.ResponseWriter, r *http.Request) {
	log.Println("get call recieved")
	// create connection
	db := l.createConnection()
	//defer close connection
	defer l.closeConnection(db)
	// read all the books
	rows, err := db.Query("select * from books")
	if err != nil {
		panic(err)
	}
	books := []Book{}
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.Id, &b.Name, &b.Isbn)
		if err != nil {
			continue
		}
		// fmt.Printf("book recived %+v\n", b)
		books = append(books, b)
	}
	// fmt.Printf("books := %+v\n", books)

	enc := json.NewEncoder(w)
	err = enc.Encode(books)
	if err != nil {
		log.Fatalf("error while getting books %s", err.Error())
	}

}

func (l library) createConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		l.host, l.port, l.user, l.password, l.dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

func (l library) closeConnection(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatalf("error while closing connection %s", err.Error())
	}
}
