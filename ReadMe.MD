**Stock**
----

Introduction
============
Http Rest API, Html and java script test page

Golang: ver 1.18

PostgresSQL: ver 14.1

Main features
- User authenticate Token support.
- Multi Goroutines support for file processing.
- Transaction support
- Dictionary result set
- Multi media log support
- Database connection pool.
- Encrypt or plane start configuration


StartApplication
============
- Configuration file

You must find config file in "./Files/ConfigPlane.json" to change postgresSQL database config on your own DB.

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
    - Return status, time, valid duration, account numbers and sing key for reuse in other services

```json
{
  "BStatus": 1,
  "Time": "2022-06-05T13:01:27.7407569+01:00",
  "Result": {
    "Username": "mohsen",
    "ValidTill": "Sun Sep 5 13:04:27 BST 2022",
    "Sign": "nzTTN4nD7U4kdar7KkHOx9wOPd8zTcoV7eDCZaiITTE/mD//E02QVIcWVno2oWIToCwZPxy/Vot/yqfNmwXr"
  }
}
```
Create transaction(deposit, withdrew)
============
- {IP}:{PORT}/transaction [POST]
    - Takes account number, amount, note, transaction type (deposit or withdrew) and sign key.
```json

```
- Response
    - Returns status and result description
```json

```
Transaction Query
============
- {IP}:{PORT}/transactions [POST]
    - Takes account number, start date, end date and sign key.

- Request
```json

```

- Response
```json

```


## By: Mohsen Taheri
## Email: taheri.mo@outlook.com
## Phone: +447759448882