package main

import(
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)


type Book struct {
	Title string
	Author string
	Isbn	string
	Borrower int
}

type Person struct {
	Name	string
	StudentId	string
	Email	string
}

func getBooks(db *sql.DB) []Book{
	//REQUIRES: db be a working connection to a mysql database with table books and fields title, author,isbn, and borrower where title and author are not null
	//ENSURES: a slice of Book structs containing data from the database will be returned
	var ret []Book
	results,_ := db.Query("SELECT title,author,isbn,borrower FROM books;")

	//This iterates over all the results and converts them to Book structs then appends it to ret
	for results.Next(){
		var nuBook Book;
		var bNum *int;
		var isbn *string;
		if err:= results.Scan(&nuBook.Title,&nuBook.Author,&isbn,&bNum); err !=nil{
			panic(err)
		}
		if isbn !=nil {
			nuBook.Isbn = *isbn;
		}
		if bNum !=nil {
			nuBook.Borrower = *bNum;
		}
		ret = append(ret,nuBook)
	}
	return ret;
}

func saveBook(b Book,db *sql.DB){
	//This function saves the book to the database
	fString := "INSERT INTO books (title,author"
	eString := fmt.Sprintf(") VALUES (%q,%q",b.Title,b.Author)
	if b.Isbn != "" {
		fString = fString + ",isbn"
		eString = eString + fmt.Sprintf(",%q",b.Isbn)
	}
	if b.Borrower != 0 {
		fString = fString + ",borrower"
		eString = eString + fmt.Sprintf("%q",b.Borrower)
	}
	query := fString + eString + ");"
	db.Query(query)

}


func main(){
	db,err := sql.Open("mysql","librarian:password@/bookshelf")
	if err != nil {
		return
	}
	defer db.Close()
	fmt.Println(getBooks(db))
}
