package API

import (
	"Github.com/mhthrh/Stock/Controller"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/DbPool"
	"Github.com/mhthrh/Stock/Utilitys/ValidationUtil"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

func RunApiOnRouter(sm *mux.Router, log *logrus.Entry, db *DbPool.DBs) {
	controller := Controller.New(log, ValidationUtil.NewValidation(), db)
	sm.Use(controller.MiddleWare)
	postR := sm.Methods(http.MethodPost).Subrouter()

	postR.HandleFunc("/signIn", controller.SignIn)
	postR.HandleFunc("/bulk", controller.Bulk)
	postR.HandleFunc("/get", controller.Get)
	postR.HandleFunc("/put", controller.Put)

	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/page", nil)
}
