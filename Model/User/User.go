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
	SignedPassword := CryptoUtil.NewKey()
	SignedPassword.Text = l.Password
	SignedPassword.Sha256()
	rows, err := PgSql.RunQuery(t.db, fmt.Sprintf("SELECT \"ID\", \"UserName\" FROM public.\"Users\" where \"UserName\"='%s' and \"Password\"='%s'", l.Username, SignedPassword.Result))
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
func (t *tool) CheckSignKey(signedKey string) ([]string, error) {

	k := CryptoUtil.NewKey()
	k.Text = signedKey
	err := k.Decrypt()
	if err != nil {
		return nil, err
	}
	spl := strings.Split(k.Result, "#")
	if len(spl) > 1 {
		return spl, nil
	}
	return nil, fmt.Errorf("error in sign key")

}
