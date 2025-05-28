package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	// Запускаем несколько рабочих горутин
	for i := 0; i < 10; i++ {
		go worker(i)
	}

	// HTTP endpoint для healthcheck (для docker-compose)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// Endpoint для ручного создания нагрузки
	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		go worker(rand.Intn(1000))
		fmt.Fprintln(w, "started extra worker")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Demo app running on :%s (pid=%d, goroutines=%d)\n", port, os.Getpid(), runtime.NumGoroutine())
	http.ListenAndServe(":"+port, nil)
}

func worker(id int) {
	for {
		// Имитируем работу: считаем что-то, иногда спим
		n := rand.Intn(1000000)
		sum := 0
		for i := 0; i < n; i++ {
			sum += i % 7
		}
		if id%3 == 0 {
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		}
	}
}