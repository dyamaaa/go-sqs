
### アクセストークンの生成
`export AUTH_TOKEN=$(openssl rand -hex 16) `などでアクセストークンを生成し環境変数AUTH_TOKENに追加


### Enqueue
```
POST /enqueue
```
```
{
    "ID": "123",
    "Payload": "Queue Message"
}
```
EX.
```
curl -X POST 'http://localhost:8080/enqueue?queueName=myQueue' \
-H 'Content-Type: application/json' \
-H 'Authorization: AUTH_TOKEN' \
-d '{"ID": "msg1", "Payload": "This is a test message"}'

```

### Dequeue
```
GET /dequeue
```
```
{
    "ID": "123",
    "Payload": "Queue Message"
}
```

Ex.
```
curl -X GET 'http://localhost:8080/dequeue?queueName=myQueue&timeout=5s' \
-H 'Authorization: AUTH_TOKEN'

```