package User

import (
	"Github.com/mhthrh/Stock/Utilitys/CryptoUtil"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/PgSql"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type User struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	LastName string    `json:"lastName"`
	UserName string    `json:"userName"`
	PassWord string    `json:"passWord"`
}
type Login struct {
	CountryCode string `json:"countryCode" validate:"required,min=2,max=3"`
	Username    string `json:"username" validate:"required,alphanum"`
	Password    string `json:"password" validate:"required,min=8,max=16"`
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

func (t *tool) SignIn(l *Login) (string, error) {
	var Countt int
	SignedPassword := CryptoUtil.NewKey()
	SignedPassword.Text = l.Password
	SignedPassword.Sha256()

	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT count(*) FROM public.country where shortname='%s'", l.CountryCode))
	if err != nil {
		return "", err
	}
	if rows.Next() {
		rows.Scan(&Countt)
	}
	if Countt != 1 {
		return "", fmt.Errorf("country not found")
	}
	rows, err = PgSql.RunQuery(t.db, fmt.Sprintf("SELECT \"ID\", \"UserName\" FROM public.\"Users\" where \"UserName\"='%s' and \"Password\"='%s'", l.Username, SignedPassword.Result))
	if err != nil {
		return "", err
	}

	if rows.Next() {
		return t.GenerateSignKey(l.CountryCode, l.Username, validationDuration), nil
	}
	return "", fmt.Errorf("user or pass invalid")

}
func (t *tool) GenerateSignKey(CountryCode, userName string, validationDuration time.Duration) string {
	j := CryptoUtil.NewKey()
	j.Text = fmt.Sprintf("%s#%s#%s", CountryCode, userName, time.Now().Add(validationDuration).Format(dateFormat))
	j.Encrypt()
	return j.Result
}
func (t *tool) CheckSignKey(user, token string) (bool, error) {
	var count int
	k := CryptoUtil.NewKey()
	k.Text = token
	err := k.Decrypt()
	if err != nil {
		return false, err
	}
	spl := strings.Split(k.Result, "#")
	if len(spl) != 3 {
		return false, fmt.Errorf("error in sign key")
	}

	rows, err := t.db.Query(fmt.Sprintf("select count(*) from public.country where shortname='%s'", spl[0]))
	defer rows.Close()
	if err != nil {
		return false, fmt.Errorf("error in sign key")
	}
	if rows.Next() {
		rows.Scan(&count)
	}
	if count != 1 {
		return false, fmt.Errorf("country not found")
	}
	if user != spl[1] {
		return false, fmt.Errorf("user not found")
	}
	rows, err = t.db.Query(fmt.Sprintf("select count(*) from  public.\"Users\" where \"UserName\"='%s'", spl[1]))
	defer rows.Close()
	if err != nil {
		return false, fmt.Errorf("error in sign key")
	}
	if rows.Next() {
		rows.Scan(&count)
	}
	if count != 1 {
		return false, fmt.Errorf("user not found")
	}
	signedTime, err := time.Parse(time.UnixDate, spl[2])
	if time.Now().Before(signedTime) {
		return true, nil
	}
	return false, fmt.Errorf("token has been expierd")

}

func (t *tool) UserDetail(token string) (Login, error) {
	k := CryptoUtil.NewKey()
	k.Text = token
	err := k.Decrypt()
	if err != nil {
		return Login{}, err
	}
	spl := strings.Split(k.Result, "#")
	if len(spl) != 3 {
		return Login{}, fmt.Errorf("parsing error")
	}
	return Login{
		CountryCode: spl[0],
		Username:    spl[1],
		Password:    "paspaspaspas",
	}, nil
}
