// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	htmlHeader = `<!DOCTYPE html>
<html>
<head>
<style type="text/css">
body { background-color:#fff; color:#333; font-family:verdana, arial, helvetica, sans-serif; font-size:13px; line-height:18px }
p,ol,ul,td { font-family: verdana, arial, helvetica, sans-serif;font-size:13px; line-height:18px}
a { color:#000 }
a:visited { color:#666 }
a:hover{ color:#fff; background-color:#000 }
tr.dir { font-weight: bold }
td.icon { font-size: 20px; }
a.icon { font-size: 27px; text-decoration: none; }
</style>
<meta charset="UTF-8" />
<title>Telegram logs</title>
</head>
<body>
<h1><a href="/">Telegram logs</a></h1>`
	htmlFooter = `</body>
</html>`
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

func httpFinishError(ctx *fasthttp.RequestCtx, err error) {
	ctx.WriteString(err.Error())
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	log.Error(err)
}

func httpFinishOK(ctx *fasthttp.RequestCtx, data string) {
	ctx.WriteString(data)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func httpRootHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	var err error

	var chats []*tgbotapi.Chat
	if chats, err = getChats(); err != nil {
		httpFinishError(ctx, err)
		return
	}

	ctx.WriteString(htmlHeader)
	ctx.WriteString(`<p>Chats:</p>
	<ul>`)

	for _, chat := range chats {
		fmt.Sprintf(`<li><a href="/chat/%d">%s</a></li>`, chat.ID, chat.Title)
	}
	ctx.WriteString("</ul>")
	ctx.WriteString(htmlFooter)
	//ctx.WriteString("OK")
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
