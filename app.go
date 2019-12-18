package main

import (
	"encoding/json"
	"fmt"
	"github.com/baagee/test-vgo/util"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"net/http"
)

type App struct {
	Router     *mux.Router //路由
	Middleware *Middleware // 中间件
	env        *Env        // 环境变量
}

// 请求的数据
type shortenRequest struct {
	Url    string `json:"url" validator:"nonzero"`
	Expire int64  `json:"expire" validator:"min=0"`
}

// 响应的结构
type shortenResponse struct {
	Short string `json:"short"`
}

func (app *App) Initialize(env *Env) {
	//打印日志时间日期和文件行号
	log.SetFlags(log.LstdFlags | log.Llongfile)
	app.Router = mux.NewRouter()
	app.env = env
	app.Middleware = &Middleware{}
	app.initRouter()
}

//初始化路由
func (app *App) initRouter() {
	m := alice.New(app.Middleware.LogHandler, app.Middleware.RecoverHandler)
	app.Router.Handle("/api/shorten", m.ThenFunc(app.createShorten)).Methods("POST")
	app.Router.Handle("/api/info", m.ThenFunc(app.getShortenInfo)).Methods("GET")
	app.Router.Handle("/{shorten:[a-zA_Z0-9]{1,11}}", m.ThenFunc(app.redirect)).Methods("GET")
}

//创建
func (app *App) createShorten(writer http.ResponseWriter, request *http.Request) {
	var req shortenRequest
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		//返回错误信息
		responseWithError(writer, util.StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("parse params failed %v", request.Body),
		})
		return
	}
	validators := validator.New()
	if err := validators.Struct(req); err != nil {
		//返回错误信息
		responseWithError(writer, util.StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("validate params failed %v", req),
		})
		return
	}
	log.Printf("%v\n", req)
	res, err := app.env.Storage.Shorten(req.Url, req.Expire)
	if err != nil {
		responseWithError(writer, err)
	} else {
		responseWithJson(writer, http.StatusCreated, shortenResponse{Short: res})
	}
}

//获取信息
func (app *App) getShortenInfo(writer http.ResponseWriter, request *http.Request) {
	vars := request.URL.Query()
	link := vars.Get("short")
	log.Println(link)
	detail, err := app.env.Storage.ShortenInfo(link)
	if err != nil {
		responseWithError(writer, err)
	} else {
		responseWithJson(writer, http.StatusOK, detail)
	}
}

// 跳转
func (app *App) redirect(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	shorten := vars["shorten"]
	log.Println(shorten)
	link, err := app.env.Storage.UnShorten(shorten)
	if err != nil {
		responseWithError(writer, err)
	} else {
		//302临时重定向
		http.Redirect(writer, request, link, http.StatusFound)
		return
	}
}

func (app *App) Run(addr string) {
	http.ListenAndServe(addr, app.Router)
}

func responseWithError(writer http.ResponseWriter, err error) {
	switch e := err.(type) {
	case util.Error:
		//自定义的Error类型
		log.Printf("request error :[%d] %s\n", e.Status(), e.Error())
		responseWithJson(writer, e.Status(), e.Error())
	default:
		log.Printf("http error : %s\n", e.Error())
		//输出500
		responseWithJson(writer, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func responseWithJson(writer http.ResponseWriter, status int, message interface{}) {
	resp, _ := json.Marshal(message)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(resp)
}
