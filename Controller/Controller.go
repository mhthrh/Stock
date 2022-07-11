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
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
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
type authenticate struct {
	User  string `json:"user"`
	Token string `json:"token"`
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
				Result.New(1001, http.StatusBadRequest, GenericError{Message: err.Error()}.Message).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return

			}
			errs := c.validation.Validate(obj1)
			if len(errs) > 0 {
				j := JsonUtil.New(nil, nil).Struct2Json(ValidationError{Messages: errs.Errors()}.Messages)
				Result.New(1002, http.StatusUnprocessableEntity, j).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), Key{}, obj1))
			next.ServeHTTP(w, r)
		case "/bulk":
			aut := authenticate{}
			json.Unmarshal([]byte(r.FormValue("aut")), &aut)
			d := c.db.Pull()
			_, err := User.New(d.Db).CheckSignKey(aut.User, aut.Token)
			if err != nil {
				Result.New(1003, http.StatusOK, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			c.db.Push(d)

			file, handler, err := r.FormFile("file")
			defer file.Close()
			if err != nil {
				Result.New(1004, http.StatusOK, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			d = c.db.Pull()
			if err := fileCheck(handler, d); err != nil {
				Result.New(1022, http.StatusAlreadyReported, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			c.db.Push(d)
			ctx := context.WithValue(r.Context(), Key{}, file)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		case "/search":
			var obj1 Stock.Search
			err := json.NewDecoder(r.Body).Decode(&obj1)
			if err != nil {
				Result.New(1005, http.StatusBadRequest, GenericError{Message: err.Error()}.Message).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			d := c.db.Pull()
			_, err = User.New(d.Db).CheckSignKey(obj1.Username, obj1.Token)
			if err != nil {
				Result.New(1006, http.StatusForbidden, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			usr, err := User.New(d.Db).UserDetail(obj1.Token)
			if err != nil {
				Result.New(1007, http.StatusForbidden, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			obj1.Usr = usr
			c.db.Push(d)
			errs := c.validation.Validate(obj1)
			if len(errs) > 0 {
				j := JsonUtil.New(nil, nil).Struct2Json(ValidationError{Messages: errs.Errors()}.Messages)
				Result.New(1008, http.StatusUnprocessableEntity, j).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), Key{}, obj1))
			next.ServeHTTP(w, r)
		case "/consume":
			var obj1 Stock.Item
			err := json.NewDecoder(r.Body).Decode(&obj1)
			if err != nil {
				Result.New(1009, http.StatusBadRequest, GenericError{Message: err.Error()}.Message).SendResponse(w)
				return
			}
			d := c.db.Pull()
			_, err = User.New(d.Db).CheckSignKey(obj1.Username, obj1.Token)
			if err != nil {
				Result.New(1010, http.StatusForbidden, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			usr, err := User.New(d.Db).UserDetail(obj1.Token)
			if err != nil {
				Result.New(1011, http.StatusForbidden, err.Error()).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			obj1.Usr = usr
			c.db.Push(d)
			errs := c.validation.Validate(obj1)
			if len(errs) > 0 {
				j := JsonUtil.New(nil, nil).Struct2Json(ValidationError{Messages: errs.Errors()}.Messages)
				Result.New(1012, http.StatusUnprocessableEntity, j).SendResponse(w)
				c.logger.Logger.Println("error ", err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), Key{}, obj1))
			next.ServeHTTP(w, r)
		default:
			Result.New(1013, http.StatusNotFound, "address not found").SendResponse(w)

		}
	})
}

func (c *Controller) SignIn(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(Key{}).(User.Login)

	d := c.db.Pull()
	response, err := User.New(d.Db).SignIn(&u)
	c.db.Push(d)
	if err != nil {
		Result.New(1014, http.StatusBadRequest, err.Error()).SendResponse(w)
		c.logger.Logger.Println("error ", err)
		return
	}
	Result.New(1, http.StatusOK, JsonUtil.New(nil, nil).Struct2Json(response)).SendResponse(w)
}
func (c *Controller) Bulk(w http.ResponseWriter, r *http.Request) {

	var Lines []Stock.Stock
	var line Stock.Stock
	var skip bool

	//csvReader := csv.NewReader(file)
	//records, err := csvReader.ReadAll()

	file := r.Context().Value(Key{}).(multipart.File)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if !skip {
			skip = !skip
			continue
		}
		spl := strings.Split(scanner.Text(), "\",\"")
		line.Country = strings.Replace(spl[0], "\"", "", -1)
		line.SKU = strings.Replace(spl[1], "\"", "", -1)
		line.Name = strings.Replace(spl[2], "\"", "", -1)
		line.Count, _ = strconv.ParseInt(strings.Replace(spl[3], "\"", "", -1), 10, 64)
		Lines = append(Lines, line)
		c.logger.Logger.Println("reading file ", line)
	}
	d := c.db.Pull()
	rr, err := Stock.New(d.Db).Bulk(Lines)
	c.db.Push(d)
	if err != nil {
		Result.New(1020, http.StatusConflict, err).SendResponse(w)
		c.logger.Logger.Println("error ", err)
		return
	}
	Result.New(1, http.StatusOK, JsonUtil.New(nil, nil).Struct2Json(rr)).SendResponse(w)
}
func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(Key{}).(Stock.Search)

	d := c.db.Pull()
	stock, err := Stock.New(d.Db).Search(&s)
	c.db.Push(d)
	if err != nil {
		Result.New(1016, http.StatusBadRequest, err.Error()).SendResponse(w)
		c.logger.Logger.Println("error ", err)
		return
	}
	Result.New(1, http.StatusOK, JsonUtil.New(nil, nil).Struct2Json(stock)).SendResponse(w)
}
func (c *Controller) Consume(w http.ResponseWriter, r *http.Request) {
	i := r.Context().Value(Key{}).(Stock.Item)

	d := c.db.Pull()
	err := Stock.New(d.Db).Consume(&i)

	if err != nil {
		Result.New(1017, http.StatusBadRequest, err.Error()).SendResponse(w)
		c.logger.Logger.Println("error ", err)
		return
	}
	Result.New(1, http.StatusOK, "success").SendResponse(w)
}

func fileCheck(h *multipart.FileHeader, d *DbPool.DB) error {
	var count int
	err := d.Db.QueryRow(fmt.Sprintf("select count(*) from public.files where name='%s'", h.Filename)).Scan(&count)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("file alredy exist")
	}
	_, err = d.Db.Exec(fmt.Sprintf("INSERT INTO public.files(id, name, size, type, datetime)VALUES ('%s', '%s', '%d', '%s', '%s')", uuid.New(), h.Filename, h.Size, h.Header, time.Now().Format(time.UnixDate)))
	if err != nil {
		return err
	}
	return nil
}
