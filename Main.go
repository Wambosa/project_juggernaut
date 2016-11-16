package main

import (

	"os"
	"fmt"
	"net"
	"time"
	"regexp"
	"os/user"
	"database/sql"
	"github.com/wambosa/easydb"
	"github.com/wambosa/jugger"
	"errors"
	"io/ioutil"
)

var (
	ConnectionString string
	BinFolder string
	BinExcludes []string
)

func main(){

	// loop ? or cron ?

	//create a mono options like handling for go
	//use the flexible switch setting tactic

	// -d 'databasename' (eventually this will be able to use polyjug.. for now just direct sqlite)
	// -u 'database username'
	// -p 'database password'
	// -bin '../bin/folder/path' (assume default 'thisFolder/bin')
	// -exclude 'Build ParcelPirate.exe, .*.txt, .*.conf, .*.db3'

	// HirelingDerp  DerpMinion will handle canned responses by reading into your profile.
	// i want it to be like a basic assistant for simple duties
	// it cant make too many assumptions (basically replace the profile scripts)
	// DerpMinion will have its own profile in the someday database so that it can retrieve pure canned responses that do not require computation

	ConnectionString = "bin/someday.db3"
	BinFolder = "./bin"

	// game idea for platformer. try to avoid dirt clouds to avoid outside smell and stay fresh in order to ask out girl.
	// get rejected if too unclean

	// start the instance and try to find a config in the database using the mac address and something else in the os package
	fmt.Println("Juggernaut Mind version 1");

	uniqueAddy, err := BuildUniqueMindAddress()

	if err != nil {fatal(uniqueAddy, err)}

	mindConfig, err := GetMindConfig(uniqueAddy)

	fmt.Println(mindConfig)

	// capability can be defined in either a file *or just the bin directory*.
	// the name of the files in the bin directory should directly correlate to an action by exact name
	// or through chain command exec text lookup

	BinExcludes = []string{
		`.*\.txt`,
		`.*\.db3`,
		`.*\.conf`,
	}

	files, err := ioutil.ReadDir(BinFolder)

	for i, file := range files {

		if file.IsDir() {continue}

		isExcluded := false

		for _, regx := range BinExcludes {

			if regexp.MustCompile(regx).MatchString(file.Name()) {
				isExcluded = true
				break
			}
		}

		if isExcluded {continue}

		fmt.Println(i, file.Name())
	}

	//once we have discovered current capabilities, compare them with existing, and add or remove as necessary.

	//after we have commited the capabilities for this mind, we will
	// SELECT * FROM ChainCommand WHERE ActionId IN(thisMind capability action Ids)

	//after the init has complete, we can query for jobs
	//SELECT * FROM Job WHERE JobStatusId = 1 AND ActionId IN(thisMind capability action ids)

	//foreach job,
	// get the job meta data
	// get all the steps for that action to complete
	// it will be necessary to replace the {0} {1} |PropertyName| with the actual values before firing off the commands
	// need to assume that the output from the previous step is going to be passed as is into the next step. this can be marked as |GENERIC_INPUT|

	// will figure out error handling as we go.

	// capture any output as the response.

	// add to the response much like the insult gen. some messages will incur a snide remark, others a general into or outro. maybe comment on the previous context of the last conversation

	// some actions will require permissions. use a whitelist type approach. assume no permissions if there are no users with permission to perform an action.
	// INSERT INTO UserPermission (ActionId, UserId)

	// the end goal is to create response records. any work that is actually done is a side effect of the various cli programs
	// maybe ResponseRecords need metadata too ? so that we can keep the pure response seprate from the sass ?
	// INSERT INTO ResponseRecord

	//from here we can wait till next loop or cron. but the database connection likely needs to be terminated until the next session.
	// NOTE: i may need to check for the database status. if the file is locked or not by another program. since there are now 3 diff programs accessing the database. up to 4 programs all in all

	fmt.Println("done")
}

func BuildUniqueMindAddress()(string, error) {

	hostname, err := os.Hostname()

	if err != nil {return "failed to get hostname", err}

	user, err := user.Current()

	ipInterfaces, err := net.Interfaces()

	if err != nil {return "failed to get network interfaces", err}

	// i am going to assume that the first interface is the primary. this could create bugs. i may need to perform some reachability test before assuming this
	if len(ipInterfaces) == 0 {
		return "", errors.New("did not find any valid network interafaces")}

	//NOTE: it is safer if i do not store the machine mac address in case of some sort of breach. the data is not necessary for uniqueness.
	//clean username because windows have an escape symbol in the username

	cleanUsername := regexp.MustCompile(`\\`).ReplaceAllString(user.Username, "_")

	uniqueId := fmt.Sprintf("%s|%s|%s", hostname, cleanUsername, ipInterfaces[0].HardwareAddr)

	return uniqueId, nil
}

func GetMindConfig(uniqueAddress string)(jugger.Mind, error){

	var mindConfig jugger.Mind

	db, err := sql.Open("sqlite3", ConnectionString)

	if err != nil {return mindConfig, err}

	defer db.Close()

	query := `
	SELECT *
	FROM Mind
	WHERE UniqueAddress LIKE '%s'
	`

	readyQuery := fmt.Sprintf(query, uniqueAddress)

	configs, err := easydb.Query(db, readyQuery)

	if err != nil {return mindConfig, errors.New("failed to query database for mind config" + readyQuery + fmt.Sprint(err))}

	if len(configs) > 0 {

		parsedTime, err := time.Parse(time.RFC3339, configs[0]["LastUpdated"].(string))

		if err != nil {return mindConfig, err}

		mindConfig = jugger.Mind {
			MindId: configs[0]["MindId"].(int),
			MindName: configs[0]["MindName"].(string),
			Nosiness: configs[0]["Nosiness"].(int),
			Sassyness: configs[0]["Sassyness"].(int),
			UniqueAddress: configs[0]["UniqueAddress"].(string),
			LastUpdated: parsedTime,
		}

		fmt.Println("found this mind configuration")

	}else{

		// if no config record exists, then create one for next time
		mindConfig = jugger.Mind{
			MindId: 0,
			MindName: "vlad",
			Nosiness: 50,
			Sassyness: 50,
			UniqueAddress: uniqueAddress,
			LastUpdated: time.Now(),
		}

		fmt.Println("Did not find existing mind configuration, so i created a new one.")
	}

	return mindConfig, nil
}

func fatal(myDescription string, err error) {
	fmt.Println("FATAL: ", myDescription)
	fmt.Println(err)
	os.Exit(1)
}