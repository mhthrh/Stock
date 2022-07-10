package Stock

import (
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/PgSql"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"sort"
	"time"
)

type Stock struct {
	id      uuid.UUID
	Country string
	Name    string
	SKU     string
	Count   int64
	Result  string
}

type Message struct {
	slice []Stock
	cnn   *sql.Tx
}
type Search struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required,base64"`
	Sku      string `json:"sku" validate:"required,alphanum"`
	Country  string `json:"country"`
}
type Item struct {
	Username string `json:"username" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Token    string `json:"token" validate:"required,base64"`
	Sku      string `json:"sku" validate:"required,alphanum"`
	Count    int    `json:"count" validate:"required,numeric,gt=0"`
}
type tool struct {
	db *sql.DB
}

var (
	validationDuration = 180 * time.Second
	dateFormat         = time.UnixDate
)

func New(db *sql.DB) *tool {
	return &tool{db: db}
}

func (t *tool) Search(s *Search) (Stock, error) {
	var result Stock
	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT  c.name, s.name, s.sku, s.count FROM public.stock s inner join  public.country c on s.country=c.shortname where s.country='%s' and s.sku='%s'", s.Country, s.Sku))
	defer rows.Close()
	if err != nil {
		return Stock{}, err
	}

	if rows.Next() {
		rows.Scan(&result.Country, &result.Name, &result.SKU, &result.Count)
		return result, nil
	}

	return Stock{}, fmt.Errorf("cant find sku")
}
func (t *tool) Put(i *Item) error {

	result, err := PgSql.ExecuteCommand(fmt.Sprintf("insert into "), t.db)
	if err != nil {
		return err
	}

	count, err := (*result).RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("not added")
	}
	return nil
}
func (t *tool) Bulk(stock []Stock) ([]Stock, error) {

	const GoRoutines = 40
	var k, j int
	step := len(stock) / GoRoutines
	j = step
	chn := make(chan Message)
	commit := true
	var t2 []Stock
	sort.SliceStable(stock, func(i, j int) bool {
		return stock[i].Count > stock[j].Count
	})

	transaction, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func(tx *sql.Tx) {
		fmt.Println("transactions ", commit)
		if !commit {
			tx.Rollback()
			return
		}
		tx.Commit()

	}(transaction)

	for i := 0; i < GoRoutines; i++ {

		go t.insert(chn)
		chn <- Message{
			slice: stock[k:j],
			cnn:   transaction,
		}
		k = j
		if j+step > len(stock) {
			j = len(stock)
		} else {
			j = j + step
		}

	}

	for i := 0; i < GoRoutines; i++ {
		result := <-chn
		t2 = append(t2, result.slice...)
	}
	return t2, nil
}
func (t *tool) insert(chn chan Message) {

	var trans *sql.Tx
	var stocks []Stock
	select {
	case msg := <-chn:
		trans = msg.cnn
		stocks = msg.slice
	}
	defer func() {
		chn <- Message{
			slice: stocks,
			cnn:   trans,
		}

	}()
	for index, s := range stocks {

		u, _ := uuid.NewUUID()
		result, err := trans.Exec(fmt.Sprintf("INSERT INTO public.stock( id, country, name, sku, Count) SELECT '%s', '%s', '%s', '%s', '%d' WHERE NOT EXISTS (SELECT id FROM public.stock where country='%s' and sku='%s')", u, s.Country, s.Name, s.SKU, s.Count, s.Country, s.SKU))
		if err != nil {
			s.Result = err.Error()
			continue
		}
		i, err := result.RowsAffected()
		if err != nil {
			s.Result = err.Error()
			continue
		}
		if i == 0 {
			ddd := fmt.Sprintf("UPDATE public.stock SET  Count=Count+'%d' WHERE country='%s' and sku='%s'", s.Count, s.Country, s.SKU)
			_, err = trans.Exec(ddd)
			if err != nil {
				s.Result = err.Error()
				continue
			}
			stocks[index].Result = "Updated"
		} else {
			stocks[index].Result = "Inserted"
		}
	}

}
