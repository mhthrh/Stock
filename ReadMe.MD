**Stock WEB API**
----

Introduction
============
Http Rest API, Html and java script test page

Golang: ver 1.18

PostgresSQL: ver 14.1

Main features
- User authenticates Token support.
- Multi Goroutines support for file processing.
- Transaction support
- Dictionary result set
- Multi-media log support
- Database connection pool.
- Encrypt or plane start configuration


StartApplication
============
- Configuration file

please find config file in "./Files/ConfigPlane.json" to change postgresSQL database config on your own DB.

- Run
    - go run ./main.go -- ip -- port
- Build
    - go build ./main.go


Services
============

WebPage
============
- Web page application
    - /page: Login

Signing
============

- {IP}:{PORT}/SignIn: Takes Country code, Username and Password with json format.
    - Request
```json
  {
  "Country": "UK",
  "username":"mohsen",
  "password":"qaz@123456"
  }
```
- Response
    - Returns status, time and token for reuse in other services, it's valid for 3 minutes.
    - Country short name, user and valid duration are in token

```json
{
  "BStatus": 1,
  "Time": "2022-07-11T06:30:26.6309626+01:00",
  "Result": "\"gtUQVsfPSHld9L/um+4yA4DuCFXANltTiP0ZHqq0M4s80/X+RyKLg8e7e9TUzgye/UJtZUQ4MfF3mDx9290J1wDm\""
}
```
Create transaction(deposit, withdrew)
============
- {IP}:{PORT}/search [POST]
    - Takes username, token and sku.
```json
{
  "username": "mohsen",
  "token": "gtUQVsfPSHld9L/um+4yA4DuCFXANltTiP0ZHqq0M4s80/X+RyKLg8e7e9TUzgye/UJtZUQ4MfF3mDx9290J1wDm",
  "sku": "fefew344ed"
}
```
- Response
    - Returns status and result description
```json
{
  "BStatus": 1,
  "Time": "2022-07-11T06:38:50.8411101+01:00",
  "Result": "{\"Country\":\"England\",\"Name\":\"Craig Ltd Soft Computer\",\"SKU\":\"dca55129ff77\",\"Count\":89,\"Result\":\"\"}"
}
```
Transaction Query
============
- {IP}:{PORT}/consume
    - Takes username, token, sku and count and check stock, if it's available in your country, reduces it in transactional method and return.

- Request
```json
{
  "username": "mohsen",
  "token": "A4rGrPmhqdqW2V+E7wL07hdQTzznZTkW4PQXmaMUBJXt0vPVYzu/e1OvEtjVs5MxXrg+kq3d+/kOvNhTqgCH0IYZ",
  "sku": "dca55129ff77",
  "Count": 5
}
```

- Response
```json
{
  "BStatus": 1,
  "Time": "2022-07-11T06:47:21.5668334+01:00",
  "Result": "success"
}
```

CSV file processing
============
- {IP}:{PORT}/bulk
  - Takes username, token, csv

- This service receives Usernames, tokens, and CSV files and starts with parallel processing with the numbers of explained Goroutines in its method. Inserting or updating data in the database is transactional for all the files, in a way, whether all file is submitted or not.
  File with the repetitive name will rise an error. Also, there is other way to receive the file like FTP.