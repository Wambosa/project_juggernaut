package main

import(
	"fmt"
	"github.com/wambosa/easydb"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func main(){

	db, err := sql.Open("sqlite3", "C:\\work\\git\\project_juggernaut\\bin\\someday.db3")

	if(err != nil){fmt.Println("FATAL:", err)}

	defer db.Close()

	records, err := easydb.RunQuery(
		"sqlite3",
		"C:\\work\\git\\project_juggernaut\\bin\\someday.db3",
		"SELECT * FROM ReceivedMessage")

	fmt.Println(records, err)

	fmt.Println("\n................................................\n")

	mySlice := []interface{} {
		"foo",
		"man",
		"shu",
	}
	fmt.Println(mySlice...)


}