package spider

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/mybee/go-spider/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type key int

const reqContexKey key = 0

// Context gospider context of each callback
type Context struct {
	Task  *Task
	Keys map[string]interface{}
	c     *colly.Collector
	nextC *colly.Collector

	ctlCtx    context.Context
	ctlCancel context.CancelFunc

	collyContext *colly.Context
	// output
	outputDB       *sql.DB
	outputCSVFiles map[string]io.WriteCloser
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (ctx *Context) Set(key string, value interface{}) {
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}
	ctx.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (ctx *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = ctx.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (ctx *Context) MustGet(key string) interface{} {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (ctx *Context) GetString(key string) (s string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (ctx *Context) GetBool(key string) (b bool) {
	if val, ok := ctx.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (ctx *Context) GetInt(key string) (i int) {
	if val, ok := ctx.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (ctx *Context) GetInt64(key string) (i64 int64) {
	if val, ok := ctx.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (ctx *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := ctx.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (ctx *Context) GetTime(key string) (t time.Time) {
	if val, ok := ctx.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (ctx *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := ctx.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (ctx *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (ctx *Context) GetStringMap(key string) (sm map[string]interface{}) {
	if val, ok := ctx.Get(key); ok && val != nil {
		sm, _ = val.(map[string]interface{})
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (ctx *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (ctx *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

func newContext(ctx context.Context, cancel context.CancelFunc, task *Task, c *colly.Collector, nextC *colly.Collector) (*Context, error) {
	gsCtx := &Context{
		Task:      task,
		c:         c,
		nextC:     nextC,
		ctlCtx:    ctx,
		ctlCancel: cancel,
	}

	if task.OutputConfig.Type == common.OutputTypeCSV {
		gsCtx.outputCSVFiles = make(map[string]io.WriteCloser)
		csvConf := task.OutputConfig.CSVConf
		if task.OutputToMultipleNamespace {
			for ns, conf := range task.MultipleNamespaceConf {
				csvname := fmt.Sprintf("%s.csv", ns)
				if err := createCSVFileIfNeeded(csvConf.CSVFilePath, csvname, conf.OutputFields); err != nil {
					return nil, err
				}
				outputPath := path.Join(csvConf.CSVFilePath, csvname)
				csvfile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
				if err != nil {
					return nil, errors.Wrapf(err, "open csv file [%s] failed", csvname)
				}
				gsCtx.outputCSVFiles[ns] = csvfile
			}
		} else {
			csvname := fmt.Sprintf("%s.csv", task.Namespace)
			if err := createCSVFileIfNeeded(csvConf.CSVFilePath, csvname, task.OutputFields); err != nil {
				return nil, err
			}
			outputPath := path.Join(csvConf.CSVFilePath, csvname)
			csvfile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return nil, errors.Wrapf(err, "open csv file [%s] failed", csvname)
			}
			gsCtx.outputCSVFiles[task.Namespace] = csvfile
		}

	}

	return gsCtx, nil
}

func (ctx *Context) cloneWithReq(req *colly.Request) *Context {
	newctx := context.WithValue(ctx.ctlCtx, reqContexKey, req)

	return &Context{
		Task:           ctx.Task,
		c:              ctx.c,
		nextC:          ctx.nextC,
		ctlCtx:         newctx,
		ctlCancel:      ctx.ctlCancel,
		outputDB:       ctx.outputDB,
		outputCSVFiles: ctx.outputCSVFiles,
	}
}

func (ctx *Context) setOutputDB(db *sql.DB) {
	ctx.outputDB = db
}

func (ctx *Context) closeCSVFileIfNeeded() {
	if len(ctx.outputCSVFiles) == 0 {
		return
	}
	for ns, closer := range ctx.outputCSVFiles {
		log.Debugf("closing csv file [%s]", ns+".csv")
		closer.Close()
	}
}

// GetRequest return the request on this context
func (ctx *Context) GetRequest() *Request {
	if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
		return newRequest(req, ctx)
	}
	return nil
}

// Retry retry current request again
func (ctx *Context) Retry() error {
	if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
		return req.Retry()
	}

	return nil
}

// PutReqContextValue sets the value for a key
func (ctx *Context) PutReqContextValue(key string, value interface{}) {
	if ctx.collyContext == nil {
		if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
			ctx.collyContext = req.Ctx
		} else {
			ctx.collyContext = colly.NewContext()
		}
	}
	ctx.collyContext.Put(key, value)
}

// GetReqContextValue return the string value for a key on ctx
func (ctx *Context) GetReqContextValue(key string) string {
	if ctx.collyContext == nil {
		if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
			ctx.collyContext = req.Ctx
		} else {
			return ""
		}
	}
	return ctx.collyContext.Get(key)
}

// GetAnyReqContextValue return the interface value for a key on ctx
func (ctx *Context) GetAnyReqContextValue(key string) interface{} {
	if ctx.collyContext == nil {
		if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
			ctx.collyContext = req.Ctx
		} else {
			return nil
		}
	}
	return ctx.collyContext.GetAny(key)
}

// Visit issues a GET to the specified URL
func (ctx *Context) Visit(URL string) error {
	return ctx.c.Visit(ctx.AbsoluteURL(URL))
}

// VisitWithContext issues a GET to the specified URL with current context
func (ctx *Context) VisitWithContext(URL string) error {
	return ctx.RequestWithContext("GET", ctx.AbsoluteURL(URL), nil, nil)
}

// VisitForNext issues a GET to the specified URL for next step
func (ctx *Context) VisitForNext(URL string) error {
	return ctx.nextC.Visit(ctx.AbsoluteURL(URL))
}

func (ctx *Context) reqContextClone() *colly.Context {
	newCtx := colly.NewContext()
	if ctx.collyContext == nil {
		return newCtx
	}

	ctx.collyContext.ForEach(func(k string, v interface{}) interface{} {
		newCtx.Put(k, v)
		return nil
	})

	return newCtx
}

// VisitForNextWithContext issues a GET to the specified URL for next step with previous context
func (ctx *Context) VisitForNextWithContext(URL string) error {
	return ctx.RequestForNextWithContext("GET", ctx.AbsoluteURL(URL), nil, nil)
}

// Post issues a POST to the specified URL
func (ctx *Context) Post(URL string, requestData map[string]string) error {
	return ctx.c.Post(ctx.AbsoluteURL(URL), requestData)
}

// PostWithContext issues a POST to the specified URL with current context
func (ctx *Context) PostWithContext(URL string, requestData map[string]string) error {
	return ctx.RequestWithContext("POST", ctx.AbsoluteURL(URL), createFormReader(requestData), nil)
}

// PostForNext issues a POST to the specified URL for next step
func (ctx *Context) PostForNext(URL string, requestData map[string]string) error {
	return ctx.nextC.Post(ctx.AbsoluteURL(URL), requestData)
}

// PostForNextWithContext issues a POST to the specified URL for next step with previous context
func (ctx *Context) PostForNextWithContext(URL string, requestData map[string]string) error {
	return ctx.RequestForNextWithContext("POST", ctx.AbsoluteURL(URL), createFormReader(requestData), nil)
}

// PostRawForNext issues a rawData POST to the specified URL
func (ctx *Context) PostRawForNext(URL string, requestData []byte) error {
	return ctx.nextC.PostRaw(ctx.AbsoluteURL(URL), requestData)
}

// PostRawForNextWithContext issues a rawData POST to the specified URL for next step with previous context
func (ctx *Context) PostRawForNextWithContext(URL string, requestData []byte) error {
	return ctx.nextC.Request("POST", ctx.AbsoluteURL(URL), bytes.NewReader(requestData), ctx.reqContextClone(), nil)
}

// Request low level method to send HTTP request
func (ctx *Context) Request(method, URL string, requestData io.Reader, hdr http.Header) error {
	return ctx.c.Request(method, URL, requestData, nil, hdr)
}

// RequestWithContext low level method to send HTTP request with context
func (ctx *Context) RequestWithContext(method, URL string, requestData io.Reader, hdr http.Header) error {
	return ctx.c.Request(method, URL, requestData, ctx.reqContextClone(), hdr)
}

// RequestForNext low level method to send HTTP request for next step
func (ctx *Context) RequestForNext(method, URL string, requestData io.Reader, hdr http.Header) error {
	return ctx.nextC.Request(method, URL, requestData, nil, hdr)
}

// RequestForNextWithContext low level method to send HTTP request for next step with previous context
func (ctx *Context) RequestForNextWithContext(method, URL string, requestData io.Reader, hdr http.Header) error {
	return ctx.nextC.Request(method, URL, requestData, ctx.reqContextClone(), hdr)
}

// PostMultipartForNext issues a multipart POST to the specified URL for next step
func (ctx *Context) PostMultipartForNext(URL string, requestData map[string][]byte) error {
	return ctx.nextC.PostMultipart(URL, requestData)
}

// SetResponseCharacterEncoding set the response charscter encoding on the request
func (ctx *Context) SetResponseCharacterEncoding(encoding string) {
	if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
		req.ResponseCharacterEncoding = encoding
	}
}

// AbsoluteURL return the absolute URL of u
func (ctx *Context) AbsoluteURL(u string) string {
	if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
		return req.AbsoluteURL(u)
	}
	return u
}

// Abort abort the current request
func (ctx *Context) Abort() {
	if req, ok := ctx.ctlCtx.Value(reqContexKey).(*colly.Request); ok {
		req.Abort()
	}
}

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}
