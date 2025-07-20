package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "vdb/internal/storage"
    "vdb/internal/models"
    "flag"
)

var db = storage.New()

func dumpDB() {
    fmt.Println(db.Dump())
}

func setHandler(w http.ResponseWriter, r *http.Request) {
    var kv models.KeyValue
    if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    err := db.Set(kv)
    if err != nil {
        http.Error(w, "Internal error", http.StatusBadRequest)
        return
    }
    fmt.Fprintf(w, "Set %s = %s\n", kv.Key.Key, kv.Value.Value)
    go dumpDB()
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    var key models.Key
    if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    val, ok := db.Get(key)
    if !ok {
        http.Error(w, "Key not found", http.StatusNotFound)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"value": val})
}

func main() {
    port := flag.Int("port", 8081, "port number")
    flag.Parse()
    db.Init()
    http.HandleFunc("/set", setHandler)
    http.HandleFunc("/get", getHandler)
    fmt.Println(fmt.Sprintf("Server running on :%d", *port))
    err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
    if err != nil {
        fmt.Println("Server error:", err)
    }
}