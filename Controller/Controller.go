package Controller

import (
	"Github.com/mhthrh/Stock/Model/Result"
	"Github.com/mhthrh/Stock/Model/Stock"
	"Github.com/mhthrh/Stock/Model/User"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/DbPool"
	"Github.com/mhthrh/Stock/Utilitys/JsonUtil"
	"Github.com/mhthrh/Stock/Utilitys/ValidationUtil"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	handler http.HandlerFunc
	obj     interface{}
	timeOut = 5000 * time.Millisecond
)

type GenericError struct {
	Message string `json:"message"`
}
type ValidationError struct {
	Messages []string `json:"messages"`
}
type Key struct{}
type Controller struct {
	logger     *logrus.Entry
	validation *ValidationUtil.Validation
	db         *DbPool.DBs
}

func New(l *logrus.Entry, v *ValidationUtil.Validation, db *DbPool.DBs) *Controller {
	return &Controller{l, v, db}
}

func (c *Controller) MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToLower(r.URL.Path) {
		case "/page":
			{
				http.ServeFile(w, r, "./Page/Index.html")
				return
			}
		case "/signin":
			var obj1 User.Login
			err := json.NewDecoder(r.Body).Decode(&obj1)
			if err != nil {
				Result.New(1003, http.StatusBadRequest, GenericError{Message: err.Error()}.Message).SendResponse(w)
				return

			}
			errs := c.validation.Validate(obj1)
			if len(errs) > 0 {
				j := JsonUtil.New(nil, nil).Struct2Json(ValidationError{Messages: errs.Errors()}.Messages)
				Result.New(1004, http.StatusUnprocessableEntity, j).SendResponse(w)
			}

			r = r.WithContext(context.WithValue(r.Context(), Key{}, obj1))
			next.ServeHTTP(w, r)
		case "/bulk":
			//obj = Stock.Stock{}
			//cnt, cancel := context.WithTimeout(context.WithValue(r.Context(), Key{}, obj), timeOut)
			//defer cancel()
			//r = r.WithContext(cnt)
			next.ServeHTTP(w, r)
		case "/2":
		}
	})
}

func (c *Controller) SignIn(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(Key{}).(User.Login)

	d := c.db.Pull()
	response, err := User.New(d.Db).SignIn(&u)
	c.db.Push(d)
	if err != nil {
		Result.New(1010, http.StatusBadRequest, err.Error()).SendResponse(w)
		return
	}
	Result.New(1, http.StatusOK, JsonUtil.New(nil, nil).Struct2Json(response)).SendResponse(w)
}

func (c *Controller) Bulk(w http.ResponseWriter, r *http.Request) {
	defer func() {
		dd := recover()
		fmt.Println(dd)
	}()
	var Lines []Stock.Stock
	var line Stock.Stock
	var skip bool
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if !skip {
			skip = !skip
			continue
		}
		spl := strings.Split(strings.Replace(scanner.Text(), "\"", "", -1), ",")
		line.Country = spl[0]
		line.SKU = spl[1]
		line.Name = spl[2]
		line.Count, _ = strconv.ParseInt(spl[3], 10, 64)
		Lines = append(Lines, line)
	}
	Stock.Bulk(Lines, c.db)

}
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(Key{}).(Stock.Search)

	d := c.db.Pull()
	stock, err := Stock.New(d.Db).Get(&s)

	if err != nil {
		Result.New(1010, http.StatusBadRequest, err.Error()).SendResponse(w)
		return
	}
	Result.New(1, http.StatusOK, JsonUtil.New(nil, nil).Struct2Json(stock)).SendResponse(w)
}
func (c *Controller) Put(w http.ResponseWriter, r *http.Request) {

}
