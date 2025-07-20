package file

import (
    "encoding/gob"
    "os"
)

const FILE string = "db_file"

func DumpMapToFile(m map[string]string) error {
    file, err := os.Create(FILE)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := gob.NewEncoder(file)
    return encoder.Encode(m)
}

func LoadMapFromFile() (map[string]string, error) {
    file, err := os.Open(FILE)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var m map[string]string
    decoder := gob.NewDecoder(file)
    err = decoder.Decode(&m)
    return m, err
}
