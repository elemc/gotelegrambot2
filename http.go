// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var (
	router = fasthttprouter.New()
)

func httpInit() {
	router.ServeFiles("/static/*filepath", options.StaticDirPath)
	router.GET("/", httpRootHandler)
	router.GET("/chat/:chat", httpChatHandler)
	router.GET("/chat/:chat/:year", httpYearHandler)
	router.GET("/chat/:chat/:year/:month", httpMonthHandler)
	router.GET("/chat/:chat/:year/:month/:day", httpDayHandler)

}

func httpServe() (err error) {
	defer wg.Done()

	httpInit()
	err = fasthttp.ListenAndServe(options.ServerAddr, router.Handler)
	return
}

func httpInitRequest(ctx *fasthttp.RequestCtx) {
	fields := make(log.Fields)
	fields["Body"] = string(ctx.PostBody())
	fields["Method"] = string(ctx.Method())

	addParamToFields := func(key, value []byte) {
		fields[string(key)] = string(value)
	}

	if string(ctx.Method()) == "POST" {
		ctx.PostArgs().VisitAll(addParamToFields)
	} else {
		ctx.QueryArgs().VisitAll(addParamToFields)
	}

	log.WithFields(fields).Debugf("Request from %s to %s", ctx.RemoteIP().String(), ctx.URI().String())
}

func httpRootHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.WriteString("OK")
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func httpChatHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.WriteString("OK")
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpYearHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.WriteString("OK")
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpMonthHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.WriteString("OK")
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpDayHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.WriteString("OK")
	ctx.SetStatusCode(fasthttp.StatusOK)
}
