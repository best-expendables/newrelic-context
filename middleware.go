package nrcontext

import (
	"fmt"
	"github.com/newrelic/go-agent"
	"net/http"
)

type TxnNameFunc func(*http.Request) string

type NewRelicMiddleware struct {
	app      newrelic.Application
	nameFunc TxnNameFunc
}

// Creates new middleware that will report time in NewRelic and set transaction in context
// It will ignore StatusNotFound, StatusBadRequest, StatusUnprocessableEntity
func NewMiddleware(appname string, license string) (*NewRelicMiddleware, error) {
	nrcfg := newrelic.NewConfig(appname, license)
	nrcfg.ErrorCollector.IgnoreStatusCodes = []int{
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusUnprocessableEntity,
	}

	app, err := newrelic.NewApplication(nrcfg)

	if err != nil {
		return nil, err
	}
	return &NewRelicMiddleware{app, makeTransactionName}, nil
}

// Same as NewMiddleware but accepts newrelic.Config
func NewMiddlewareWithConfig(c newrelic.Config) (*NewRelicMiddleware, error) {
	app, err := newrelic.NewApplication(c)
	if err != nil {
		return nil, err
	}
	return &NewRelicMiddleware{app, makeTransactionName}, nil
}

// Same as NewMiddleware but accepts newrelic.Application
func NewMiddlewareWithApp(app newrelic.Application) *NewRelicMiddleware {
	return &NewRelicMiddleware{app, makeTransactionName}
}

// Allows to change transaction name. By default `fmt.Sprintf("%s %s", r.Method, r.URL.Path)`
func (nr *NewRelicMiddleware) SetTxnNameFunc(fn TxnNameFunc) {
	nr.nameFunc = fn
}

// Creates wrapper for your handler
func (nr *NewRelicMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := nr.app.StartTransaction(nr.nameFunc(r), w, r)
		r = r.WithContext(ContextWithTxn(r.Context(), txn))
		defer txn.End()
		h.ServeHTTP(txn, r)
	})
}

func makeTransactionName(r *http.Request) string {
	return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
}
