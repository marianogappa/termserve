package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	port := 8080
	if len(os.Args) >= 3 {
		port, _ = strconv.Atoi(os.Args[2])
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/js_helper", jsHelperHandler(port))
	http.HandleFunc("/backend", backendHandler)
	// if len(os.Args >= 3) {
	// 	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	// }

	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, os.Args[1])
}

func jsHelperHandler(port int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		jsDeps := fmt.Sprintf(`function run(cmd) {
		return fetch(
			'http://localhost:%v/backend',
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
`, port)
		w.Header().Add("Content-Type", "text/javascript")
		fmt.Fprint(w, jsDeps)
	}
}

type response struct {
	Stdout string `json:"stdout"`
}

func backendHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(400)
		return
	}
	out, err := exec.Command("bash", "-c", string(body)).Output()
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error running command [%v]: %v", string(body), err)
		return
	}
	log.Printf("Successfully ran command [%v]", string(body))
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Stdout: string(out)})
}
