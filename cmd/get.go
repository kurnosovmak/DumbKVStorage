package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "vdb/internal/models"
)

func getValue(url, key string) (models.Value, error) {
    data := map[string]string{
        "key":   key,
    }
    jsonData, err := json.Marshal(data)
    if err != nil {
        return models.Value{}, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return models.Value{}, err
    }
    if resp.StatusCode != http.StatusOK {
        return models.Value{}, fmt.Errorf("server returned status: %s", resp.Status)
    }
    defer resp.Body.Close()

    var val models.Value
    err = json.NewDecoder(resp.Body).Decode(&val)
    if err != nil {
        return models.Value{}, err
    }

    return val, nil
}

func main() {
    var k string
    fmt.Println("Write key")
    fmt.Scanf("%s", &k)
    url := "http://localhost:8081/get"
    val, err := getValue(url, k)
    if err != nil {
        fmt.Println("Error:", err)
    } else if len(val.Value) == 0 {
        fmt.Println("Value doesnt exists")
    } else {
        fmt.Println("Value:", val.Value)
    }
}