// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func httpFinish(ctx *fasthttp.RequestCtx, status int, data string) {
	ctx.WriteString(data)
	ctx.SetStatusCode(status)
}

func httpFinishError(ctx *fasthttp.RequestCtx, err error) {
	httpFinish(ctx, fasthttp.StatusInternalServerError, err.Error())
	log.Error(err)
}

func httpFinishBadParam(ctx *fasthttp.RequestCtx, data string) {
	httpFinish(ctx, fasthttp.StatusBadRequest, data)
	log.Warnf(data)
}

func httpFinishOK(ctx *fasthttp.RequestCtx, data string) {
	httpFinish(ctx, fasthttp.StatusOK, data)
	log.Debugf("Status OK: %s", data)
}

func writeStringList(ctx *fasthttp.RequestCtx, name string, list []string) {
	if len(list) > 0 {
		ctx.WriteString(fmt.Sprintf("<h2>%s:</h2>\n<ul>", name))
		ctx.WriteString(strings.Join(list, "\n"))
		ctx.WriteString("</ul>")
	}
}

func writeTable(ctx *fasthttp.RequestCtx, path string, data []string, columns int) {
	currentColumn := 0
	ctx.WriteString("\n<table>\n\t<tr>\n")
	for _, d := range data {
		currentColumn++
		if currentColumn > columns {
			currentColumn = 1
			ctx.WriteString("\t</tr>\n\t<tr>\n")
		}
		ctx.WriteString(fmt.Sprintf("\t\t<td>\n\t\t\t<a href='%s/%s'>%s</a>\n\t\t</td>\n", path, d, d))
	}
	for i := currentColumn + 1; i <= columns; i++ {
		ctx.WriteString("\t\t<td></td>\n")
	}
	ctx.WriteString("\t</tr>\n</table>\n")
}

func httpRootHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	ctx.SetContentType("text/html")

	var (
		err          error
		chats        []tgbotapi.Chat
		privateChats []string
		groups       []string
		channels     []string
	)
	if chats, err = getChats(); err != nil {
		httpFinishError(ctx, err)
		return
	}

	for _, chat := range chats {
		if chat.IsGroup() || chat.IsSuperGroup() {
			groups = append(groups, fmt.Sprintf(`<li><a href="/chat/%d">%s</a></li>`, chat.ID, chat.Title))
		} else if chat.IsPrivate() {
			chatName := ""
			if chat.UserName == "" {
				chatName = strings.TrimSpace(fmt.Sprintf("%s %s", chat.FirstName, chat.LastName))
			} else if chat.FirstName != "" || chat.LastName != "" {
				chatName = strings.TrimSpace(fmt.Sprintf("%s (%s %s)", chat.UserName, chat.FirstName, chat.LastName))
			}
			privateChats = append(privateChats, fmt.Sprintf(`<li><a href="/chat/%d">%s</a></li>`, chat.ID, chatName))
		} else if chat.IsChannel() {
			channels = append(channels, fmt.Sprintf(`<li><a href="/chat/%d">%s</a></li>`, chat.ID, chat.Title))
		}
	}

	ctx.WriteString(htmlHeader)
	writeStringList(ctx, "Groups", groups)
	writeStringList(ctx, "Users", privateChats)
	writeStringList(ctx, "Channels", channels)
	ctx.WriteString(htmlFooter)

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func httpChatHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	var (
		err    error
		years  []string
		chatID int64
	)

	strChatID := ctx.UserValue("chat").(string)
	if chatID, err = strconv.ParseInt(strChatID, 10, 64); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Chat ID is not integer"))
		log.Error(err)
		return
	}

	ctx.SetContentType("text/html")
	ctx.WriteString(htmlHeader)

	if years, err = getChatYears(chatID); err != nil {
		httpFinishError(ctx, err)
		return
	}

	if len(years) == 0 {
		httpFinishOK(ctx, htmlFooter)
		return
	}

	ctx.WriteString("<h2>Years:</h2>")
	writeTable(ctx, fmt.Sprintf("/chat/%d", chatID), years, 3)
	ctx.WriteString(htmlFooter)
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpYearHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	var (
		err    error
		months []string
		chatID int64
		year   int
	)

	strChatID := ctx.UserValue("chat").(string)
	if chatID, err = strconv.ParseInt(strChatID, 10, 64); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Chat ID is not integer"))
		log.Error(err)
		return
	}
	strYear := ctx.UserValue("year").(string)
	if year, err = strconv.Atoi(strYear); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Year is not integer"))
		log.Error(err)
		return
	}

	ctx.SetContentType("text/html")
	ctx.WriteString(htmlHeader)

	if months, err = getChatMonths(chatID, year); err != nil {
		httpFinishError(ctx, err)
		return
	}

	if len(months) == 0 {
		httpFinishOK(ctx, htmlFooter)
		return
	}

	ctx.WriteString("<h2>Months:</h2>")
	writeTable(ctx, fmt.Sprintf("/chat/%d/%d", chatID, year), months, 4)
	ctx.WriteString(htmlFooter)
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpMonthHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	var (
		err    error
		days   []string
		chatID int64
		year   int
		month  int
	)

	strChatID := ctx.UserValue("chat").(string)
	if chatID, err = strconv.ParseInt(strChatID, 10, 64); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Chat ID is not integer"))
		log.Error(err)
		return
	}
	strYear := ctx.UserValue("year").(string)
	if year, err = strconv.Atoi(strYear); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Year is not integer"))
		log.Error(err)
		return
	}
	strMonth := ctx.UserValue("month").(string)
	if month, err = strconv.Atoi(strMonth); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Month is not integer"))
		log.Error(err)
		return
	}

	ctx.SetContentType("text/html")
	ctx.WriteString(htmlHeader)

	if days, err = getChatDays(chatID, year, month); err != nil {
		httpFinishError(ctx, err)
		return
	}

	if len(days) == 0 {
		httpFinishOK(ctx, htmlFooter)
		return
	}

	ctx.WriteString("<h2>Days:</h2>")
	writeTable(ctx, fmt.Sprintf("/chat/%d/%d/%d", chatID, year, month), days, 5)
	ctx.WriteString(htmlFooter)
	ctx.SetStatusCode(fasthttp.StatusOK)
}
func httpDayHandler(ctx *fasthttp.RequestCtx) {
	httpInitRequest(ctx)
	var (
		err    error
		msgs   []Message
		chatID int64
		year   int
		month  int
		day    int
	)

	strChatID := ctx.UserValue("chat").(string)
	if chatID, err = strconv.ParseInt(strChatID, 10, 64); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Chat ID is not integer"))
		log.Error(err)
		return
	}
	strYear := ctx.UserValue("year").(string)
	if year, err = strconv.Atoi(strYear); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Year is not integer"))
		log.Error(err)
		return
	}
	strMonth := ctx.UserValue("month").(string)
	if month, err = strconv.Atoi(strMonth); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Month is not integer"))
		log.Error(err)
		return
	}
	strDay := ctx.UserValue("day").(string)
	if day, err = strconv.Atoi(strDay); err != nil {
		httpFinishBadParam(ctx, fmt.Sprintf("Day is not integer"))
		log.Error(err)
		return
	}

	ctx.SetContentType("text/html")
	ctx.WriteString(htmlHeader)

	if msgs, err = getMessages(chatID, year, month, day); err != nil {
		httpFinishError(ctx, err)
		return
	}

	if len(msgs) == 0 {
		httpFinishOK(ctx, htmlFooter)
		return
	}

	ctx.WriteString("<h2>Messages:</h2>")

	ctx.WriteString(`<table width="80%">
<thead>
	<tr>
		<th align="center" width="5%"></th>
		<th align="center width="10%">Time</th>
		<th align="center" width="10%">User</th>
		<th align="center" width="75%">Message</th>
	</tr>
</thead>
<tbody>`)

	var data []string
	for _, msg := range msgs {
		messageTime := time.Unix(int64(msg.Date), 0).Format("15:04:05")
		user := msg.UserFrom.String()
		messageText := html.EscapeString(msg.Text)
		re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
		messageText = re.ReplaceAllString(messageText, `<a href="$0">$0</a>`)
		messageText = strings.Replace(messageText, "\n", "<br/>")

		if msg.ReplyToMessage != nil {
			lt := time.Unix(int64(msg.ReplyToMessage.Date), 0)
			replyLink := fmt.Sprintf("/chat/%d/%d/%d/%d#%s", msg.Chat.ID, lt.Year(), lt.Month(), lt.Day(), lt.Format("15:04:05"))
			messageText = fmt.Sprintf(`<p class="reply"> <a href="%s">></a> %s</p><p>%s</p>`, replyLink, html.EscapeString(msg.ReplyToMessage.Text), messageText)
		}

		photo, _ := getUserPhotoFilename(msg.UserFrom)
		if msg.Audio != nil {
			messageText += fmt.Sprintf(`<p><a href="/static/%s">Audio in message</a></p>`, getShortFileName(msg.Audio.FileID))
		}
		if msg.Document != nil {
			messageText += fmt.Sprintf(`<p><a href="/static/%s">Document in message</a></p>`, getShortFileName(msg.Document.FileID))
		}
		if msg.Photo != nil {
			f := (*msg.Photo)[len(*msg.Photo)-1]
			messageText += "<p>"
			messageText += fmt.Sprintf(`<p><a href="/static/%s"><img src="/static/%s"></img></a>`, getShortFileName(f.FileID), getShortFileName(f.FileID))
			messageText += "</p>"
		}
		if msg.Sticker != nil {
			messageText += fmt.Sprintf(`<p><img src="/static/%s"></img></p>`, getShortFileName(msg.Sticker.FileID))
		}
		if msg.Video != nil {
			messageText += fmt.Sprintf(`<p><a href="/static/%s">Video in message</a></p>`, getShortFileName(msg.Video.FileID))
		}
		if msg.Voice != nil {
			messageText += fmt.Sprintf(`<p><a href="/static/%s">Voice in message</a></p>`, getShortFileName(msg.Voice.FileID))
		}

		photoTD := fmt.Sprintf(`<td align="center"><a href="/static/%s"><img src="/static/%s" height="30px" width="30px"></img></td>`, photo, photo)
		if photo == "" {
			photoTD = `<td align="center">no image</td>`
		}

		data = append(data, (fmt.Sprintf(`	<tr style="background-color: #F5F5F5;">
		%s
		<td align="center"><a href="#%s" id="%s">%s</a></td>
		<td><strong>%s</strong></td>
		<td>%s</td>
	<tr>`, photoTD, messageTime, messageTime, messageTime, html.EscapeString(user), messageText)))
	}
	ctx.WriteString(strings.Join(data, "\n"))

	ctx.WriteString("</tbody>\n</table>")
	ctx.WriteString(htmlFooter)
	ctx.SetStatusCode(fasthttp.StatusOK)
}
