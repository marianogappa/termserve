package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/js_helper", jsHelperHandler)
	http.HandleFunc("/backend", backendHandler)
	// if len(os.Args >= 3) {
	// 	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	// }
	http.ListenAndServe(":8080", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, os.Args[1])
}

func jsHelperHandler(w http.ResponseWriter, r *http.Request) {
	jsDeps := `function run(cmd) {
		return fetch(
			'http://localhost:8080/backend',
			{
				method: 'POST',
				body: cmd,
				headers: {
					'Content-Type': 'application/json'
				},
			}
		)
		.then(response => response.json())
		.then(response => response.stdout)
	}

	function $(q, v) {
		document.querySelector(q).innerHTML = v
	}
`
	w.Header().Add("Content-Type", "text/javascript")
	fmt.Fprint(w, jsDeps)
}

type response struct {
	Stdout string `json:"stdout"`
}

func backendHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(400)
	}
	out, err := exec.Command("bash", "-c", string(body)).Output()
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error running command [%v]: %v", string(body), err)
	}
	log.Printf("Successfully ran command [%v]: %v", string(body), string(out))
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Stdout: string(out)})
}
