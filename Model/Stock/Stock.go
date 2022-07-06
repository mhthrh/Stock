package Stock

import (
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/DbPool"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/PgSql"
	"context"
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
}

type Search struct {
	token string `json:"token" validate:"required,alphanum"`
	Sku   string `json:"sku" validate:"required,alphanum"`
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

func (t *tool) Get(s *Search) (Stock, error) {
	var st Stock
	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT \"ID\", \"UserName\" FROM public.\"Users\" where \"UserName\"='%s' and \"Password\"='%s'", s.Sku, s.token))
	if err != nil {
		return Stock{}, err
	}

	if rows.Next() {
		rows.Scan(&st.Name, &st.SKU, &st.Country)
		return st, nil
	}

	return Stock{}, fmt.Errorf("cant find sku")
}
func (t *tool) Put(s *Search) (Stock, error) {
	var st Stock
	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT \"ID\", \"UserName\" FROM public.\"Users\" where \"UserName\"='%s' and \"Password\"='%s'", s.Sku, s.token))
	if err != nil {
		return Stock{}, err
	}

	if rows.Next() {
		rows.Scan(&st.Name, &st.SKU, &st.Country)
		return st, nil
	}

	return Stock{}, fmt.Errorf("cant find sku")
}

func Bulk(stock []Stock, db *DbPool.DBs) {
	defer func() {
		d := recover()
		fmt.Println(d)
	}()
	const thread = 10
	//var ctx context.Context
	chn := make(chan bool, 10000)
	commit := true

	sort.SliceStable(stock, func(i, j int) bool {
		return stock[i].Count > stock[j].Count
	})
	cnn := db.Pull()
	transaction, err := cnn.Db.Begin()
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		if !commit {
			tx.Rollback()
			return
		}
		tx.Commit()

	}(transaction)

	ctx, _ := context.WithTimeout(context.WithValue(context.Background(), 1, stock[0:]), time.Second*25)
	insert(transaction, ctx, 1, &chn)

	select {
	case <-ctx.Done():
		commit = false
	case <-chn:
		commit = true
	}

}

func insert(t *sql.Tx, c context.Context, i int, chn *chan bool) {
	stock := c.Value(i).([]Stock)
	count := 0
	for _, s := range stock {

		rows := t.QueryRow(fmt.Sprintf("SELECT Count(*) FROM public.stock where country='%s' and sku='%s'", s.Country, s.SKU))
		if err := rows.Scan(&count); err != nil {
			c.Done()
		}
		if count == 0 {
			u, _ := uuid.NewUUID()
			r, err := t.Exec(fmt.Sprintf("INSERT INTO public.stock( id, country, name, sku, Count) VALUES ('%s', '%s', '%s', '%s', '%d')", u, s.Country, s.Name, s.SKU, s.Count))
			if err != nil {
				c.Done()
			}
			i, err := r.RowsAffected()
			if err != nil {
				c.Done()

			}
			if i != 1 {
				c.Done()
			}
			continue
		}
		r, err := t.Exec(fmt.Sprintf("UPDATE public.stock SET  Count=Count+'%d' WHERE country='%s' and sku='%s'", s.Count, s.Country, s.SKU))
		if err != nil {
			c.Done()
			return
		}

		i, err := r.RowsAffected()
		if err != nil {
			c.Done()
			return
		}
		if i != 1 {
			c.Done()
			return
		}
	}
	*chn <- true
}
