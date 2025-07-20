# DumbKVStorage

Run server:
```
chmod +x run_server.sh
./run_server.sh 8081
```

Set key:
```
curl -X POST -H "Content-Type: application/json" \
    -d '{"key":"yourKey","value":"yourValue"}' \
    http://localhost:8081/set
```

Get key:
```
curl -X POST -H "Content-Type: application/json" \
    -d '{"key":"yourKey"}' \
    http://localhost:8081/get
```