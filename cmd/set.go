package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "testdb/internal/models"
)

func setKeyValue(url string, kv models.KeyValue) error {
    data := map[string]string{
        "key":   kv.Key.Key,
        "value": kv.Value.Value,
    }
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server returned status: %s", resp.Status)
    }
    return nil
}

func main() {
    var k,v string
    fmt.Println("Set key and value as: \"Key Value\"")
    fmt.Scanf("%s %s", &k, &v)
    url := "http://localhost:8081/set" // Change to your server's endpoint
    kv := models.KeyValue{
        Key:   models.Key{Key: k},
        Value: models.Value{Value: v},
    }
    err := setKeyValue(url, kv)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Key-value set successfully!")
    }
}