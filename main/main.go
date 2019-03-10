package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)


/*
 TODO: use json unmarshall


*/

var myClient = &http.Client{Timeout: 10 * time.Second}


/*
type Foo struct {
	Bar string
}

func main() {
	foo1 := new(Foo) // or &Foo{}
	getJson("http://example.com", foo1)
	println(foo1.Bar)

	// alternately:

	foo2 := Foo{}
	getJson("http://example.com", &foo2)
	println(foo2.Bar)
}

*/

type Response struct {
	string string
}

func main() {
	fmt.Printf("hello, world\n")

	myObject := new(Response)

	response, err := myClient.Get("https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=30")
	if err != nil {
		fmt.Printf("there was an error\n")
	}
	if err == nil {
		fmt.Printf("there wasnt an error\n")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		fmt.Printf("status  ok\n")
		json.NewDecoder(response.Body).Decode(myObject)
		println(myObject.string)


	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("status not ok\n")
	}
}

/*

package main

import (
"encoding/json"
"fmt"
"log"
"net"
"net/http"
)

func main() {
	http.DefaultServeMux.HandleFunc("/x.json", jsonHandler)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(l, nil)

	baseURL := "http://" + l.Addr().String()
	type result struct {
		Foo int
	}

	tests := []struct {
		url    string
		result interface{}
	}{{
		url:    baseURL + "/",
		result: new(result),
	}, {
		url:    baseURL + "/x.json",
		result: nil,
	}, {
		url:    baseURL + "/x.json",
		result: new(result),
	}}
	for i, test := range tests {
		err := getJSON(test.url, test.result)
		if err != nil {
			fmt.Printf("test %d: error %v\n", i, err)
		} else {
			fmt.Printf("test %d: ok with result %#v\n", i, test.result)
		}
	}
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"Foo": 1234}`))
}

// getJSON fetches the contents of the given URL
// and decodes it as JSON into the given result,
// which should be a pointer to the expected data.
func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot fetch URL %q: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}
	// We could also check the resulting content type
	// here too.
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("cannot decode JSON: %v", err)
	}
	return nil
}

*/