package notes
import (
	"fmt"
	"reflect"
)

func PrintMap(aMap map[string]interface{}){

	fmt.Println("TYPE:", reflect.TypeOf(aMap))

	for key, val := range aMap{
		fmt.Println("key:", key, "val:", val, "\n")
	}
}


