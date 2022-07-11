package Stock

import (
	"Github.com/mhthrh/Stock/Model/User"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/PgSql"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"sort"
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
	Usr      User.Login
}
type Item struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required,base64"`
	Sku      string `json:"sku" validate:"required,alphanum"`
	Count    int    `json:"count" validate:"required,numeric,gt=0"`
	Usr      User.Login
}
type tool struct {
	db *sql.DB
}

func New(db *sql.DB) *tool {
	return &tool{db: db}
}

func (t *tool) Search(s *Search) (Stock, error) {
	var result Stock
	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT c.name,s.name,s.sku,s.count FROM public.stock s inner join  public.country c on s.country=c.shortname where s.country='%s' and s.sku='%s'", s.Usr.CountryCode, s.Sku))
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
func (t *tool) Consume(i *Item) error {
	commit := false
	count := 0
	transaction, err := t.db.Begin()

	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if !commit {
			tx.Rollback()
			return
		}
		tx.Commit()
	}(transaction)

	rows, err := transaction.Query(fmt.Sprintf("SELECT count FROM public.stock where country ='%s' and sku='%s'", i.Usr.CountryCode, i.Sku))
	if err != nil {
		commit = false
		return err
	}
	if rows.Next() {
		rows.Scan(&count)
	}
	if count-i.Count < 0 {
		return fmt.Errorf("count is more than stock")
	}
	result, err := transaction.Exec(fmt.Sprintf("INSERT INTO public.market(id, userid, stockid, count, datetime) select gen_random_uuid(),(SELECT \"ID\" FROM public.\"Users\" where \"UserName\"='%s' ),(select  id from public.stock where country='%s' and sku='%s'),'%d',CURRENT_TIMESTAMP", i.Usr.Username, i.Usr.CountryCode, i.Sku, i.Count))
	if err != nil {
		commit = false
		return err
	}
	cnt, _ := result.RowsAffected()
	if cnt != 1 {
		commit = false
		return fmt.Errorf("cant insert to table")
	}
	result, err = transaction.Exec(fmt.Sprintf("UPDATE public.stock SET  count=count-'%d' WHERE country='%s' and sku='%s' and count-'%d'>0", i.Count, i.Usr.CountryCode, i.Sku, i.Count))
	if err != nil {
		commit = false
		return err
	}
	cnt, _ = result.RowsAffected()
	if cnt != 1 {
		commit = false
		return fmt.Errorf("error in update")
	}

	commit = true
	return nil
}
func (t *tool) Bulk(stock []Stock) ([]Stock, error) {
	const GoRoutines = 1
	var k, j int
	step := len(stock) / GoRoutines
	j = step
	chn := make(chan Message)
	commit := true
	var resultStock []Stock
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
		resultStock = append(resultStock, result.slice...)
	}

	return resultStock, nil
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

		fmt.Println("insert to db ", s)

		ddd := fmt.Sprintf("UPDATE public.stock SET  Count=Count+'%d' WHERE country='%s' and sku='%s' and count+'%d'>=0", s.Count, s.Country, s.SKU, s.Count)
		result, err := trans.Exec(ddd)
		if err != nil {
			s.Result = err.Error()
			continue
		}
		i, err := result.RowsAffected()

		if i == 0 {
			if s.Count <= 0 {
				s.Result = "cannot be less than zero"
				continue
			}
			u, _ := uuid.NewUUID()
			result, err := trans.Exec(fmt.Sprintf("INSERT INTO public.stock( id, country, name, sku, Count) SELECT '%s', '%s', '%s', '%s', '%d' WHERE NOT EXISTS (SELECT id FROM public.stock where country='%s' and sku='%s')", u, s.Country, s.Name, s.SKU, s.Count, s.Country, s.SKU))
			if err != nil {
				s.Result = err.Error()
				continue
			}
			i, err = result.RowsAffected()
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
