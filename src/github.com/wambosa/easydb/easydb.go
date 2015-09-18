package easydb

import (
	"errors"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func HasDriver(request string)(bool){
	importedDrivers := map[string]bool {
		"sqlite3": true,
	}
	return importedDrivers[request]
}

func RunQuery(driverName string, connString string, query string)([]map[string]interface{}, error){

	if !HasDriver(driverName) {
		return nil, errors.New("No Support for driver "+ driverName)}

	db, err := sql.Open(driverName, connString)

	if err != nil {return nil, err}

	defer db.Close()

	return Query(db, query)
}

func RunExec(driverName string, connString string, query string, args ...interface{})(sql.Result ,error){

	if !HasDriver(driverName) {
		return nil, errors.New("No Support for driver "+ driverName)}

	db, err := sql.Open(driverName, connString)

	if(err != nil){return nil, err}

	defer db.Close()

	return Exec(db, query, args...)
}

func Query(db *sql.DB, query string)([]map[string]interface{}, error){

	rows, err := db.Query(query)

	if(err != nil){return nil, err}

	records := make([]map[string]interface{}, 0)

	headers, _ := rows.Columns()

	defer rows.Close()

	aRecordValues := make([]interface{}, len(headers))
	aRecordValuesPtrs := make([]interface{}, len(headers))

	for rows.Next() {

		for i, _ := range headers {
			aRecordValuesPtrs[i] = &aRecordValues[i]}

		rows.Scan(aRecordValuesPtrs...)

		thisRecord := make(map[string]interface{})

		for i, fieldName := range headers {
			thisRecord[fieldName] = aRecordValues[i]}

		records = append(records, thisRecord)
	}

	return records, rows.Err()
}

func Exec(db *sql.DB, query string, args ...interface{})(sql.Result, error){

	transaction, err := db.Begin()

	if(err != nil){return nil, err}

	statement, err := transaction.Prepare(query)

	if(err != nil){return nil, err}

	defer statement.Close()

	res, err := statement.Exec(args...)

	if(err != nil){return res, err}

	transaction.Commit()

	return res, nil
}

func getColumnHeadersAndRows(database *sql.DB, query string)([]string, [][]interface{}, error){

	//this one would split out the headers and the data into two seperate returns to be combined later. maybe do this...
	return nil,nil,nil
}