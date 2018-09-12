package main

import(
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"regexp"
	"encoding/hex"
	"os"
	"golang.org/x/crypto/ssh/terminal"
	"bufio"
)


type Book struct {
	Title string
	Author string
	Isbn	string
	Borrower string
	Owner	string
}

type Person struct {
	Name	string
	StudentId	string
	Email	string
	Username	string
	Passhash	[]byte
	Salt	[]byte
}

func getBooks(db *sql.DB) []Book{
	//REQUIRES: db be a working connection to a mysql database with table books and fields title, author,isbn, and borrower where title and author are not null
	//ENSURES: a slice of Book structs containing data from the database will be returned
	var ret []Book
	results,_ := db.Query("SELECT title,author,isbn,borrower,owner FROM books;")

	//This iterates over all the results and converts them to Book structs then appends it to ret
	for results.Next(){
		var nuBook Book;
		var bNum *string;
		var isbn *string;
		var own *string;
		if err:= results.Scan(&nuBook.Title,&nuBook.Author,&isbn,&bNum,&own); err !=nil{
			panic(err)
		}
		if isbn !=nil {
			nuBook.Isbn = *isbn;
		}
		if bNum !=nil {
			nuBook.Borrower = *bNum;
		}
		if own !=nil{
			nuBook.Owner = *own;
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
	if b.Borrower != "" {
		fString = fString + ",borrower"
		eString = eString + fmt.Sprintf("%q",b.Borrower)
	}
	query := fString + eString + ");"
	db.Query(query)

}

func bytesToDB(by []byte) string{
	//Combines all of the bytes in a str
	ret := fmt.Sprintf("%x",by)
	process := regexp.MustCompile("\\s|\\[|\\]")
	return process.ReplaceAllString(ret,"")
}

func dbToBytes(str string) []byte{
	ret := make([]byte, 64)
	box,_ := hex.Decode(ret,[]byte(str))
	return ret[:box]
}

func addUser(db *sql.DB,user Person) (Person,error){
	//REQUIRE: person have non-null fields except for dbid and salt. Also pashash should be set to their password.
	//ensures: database has the new user and a updated person is returned
	salt := make([]byte,32)
	 io.ReadFull(rand.Reader,salt)
	hash, _ := scrypt.Key(user.Passhash,salt,1<<14,8,1,64)
	fmt.Printf("gen %x\n",(hash))
	query := fmt.Sprintf("INSERT INTO people(name,studentId,email,username,passhash,salt) VALUES(%q,%q,%q,%q,%q,%q);",user.Name,user.StudentId,user.Email,user.Username,bytesToDB(hash),bytesToDB(salt))
	_,err := db.Query(query)
	fmt.Println(err)
	ret := Person{user.Name,user.StudentId,user.Email,user.Username,hash,salt}
	return ret,err;
}

func login(db *sql.DB,user Person) (Person,bool){
	query := fmt.Sprintf("SELECT passhash,salt from people where username=%q;",user.Username)
	res,_ := db.Query(query)
	var salt string;
	var pHash string;
	if res.Next(){
		res.Scan(&pHash,&salt)
	} else {
		return user,false
	}
	unsalt := dbToBytes(salt)
	unhash := dbToBytes(pHash)
	hash, _ := scrypt.Key(user.Passhash,unsalt,1<<14,8,1,64)
	//fmt.Printf("given: %v\ngenerated: %v\n",hash,[]byte(pHash))
	fmt.Printf("hash %x\n", unhash)
	if (string(hash)==string( unhash)) {
		return Person{user.Name,user.StudentId,user.Email,user.Username,unhash,unsalt},true
	}
	return user,false
}


func main(){
	conf,err := os.Open("/etc/librarian.conf")
	librarian := ""
	libpasswd := ""
	basename := ""
	if err != nil {
		fmt.Println("Attempting first time set up")
		conf,err = os.Create("/etc/librarian.conf")
		if err != nil {
			fmt.Printf("FAILED TO CREATE CONFIGURATION FILE\n%s",err)
			return
		}
		fmt.Println("Enter Database name")
		fmt.Scanln(&basename)
		fmt.Println("Enter database username")
		fmt.Scanln(&librarian)
		fmt.Printf("Password: ")
		passwd,err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Println("Something Went wrong")
			return
		}
		libpasswd = string(passwd)
		fmt.Fprintf(conf,"%s\n%s\n%s\n",librarian,libpasswd,basename)
	} else {
		configReader := bufio.NewReader(conf)
		librarian,_ = configReader.ReadString('\n')
		libpasswd,_ = configReader.ReadString('\n')
		basename,_ = configReader.ReadString('\n')
	}
	conf.Close()

	db,err := sql.Open("mysql",librarian + ":"+ libpasswd +"@/"+ basename)
	if err != nil {
		return
	}
	defer db.Close()
	fmt.Println(err)
}

