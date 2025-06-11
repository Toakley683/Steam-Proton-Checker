package main

import (
	"fmt"
	"log"
)

func main() {

	APIKey, err := GetAPIKey()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(APIKey.Key)

}
