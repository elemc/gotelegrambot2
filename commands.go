// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

func commandsMainHandler(msg *tgbotapi.Message) {
	cmd := msg.Command()
	args := msg.CommandArguments()
	log.Debugf("Command from %s: `%s %s`", msg.From.String(), cmd, args)
	switch strings.ToLower(cmd) {
	case "start":
		go commandsStartHandler(msg)
	case "ban", "unban":
		go commandsBanHandler(msg)
	case "dnf", "yum":
		go commandsDNFHandler(msg)
	case "flood":
		go commandsFloodHandler(msg)
	case "ping":
		go commandsPingHandler(msg)
	case "help":
		go commandsHelpHandler(msg)
	case "pid":
		go commandsPIDHandler(msg)
	case "link":
		go commandsLinkHandler(msg)
	default:

	}
}

func commandsStartHandler(msg *tgbotapi.Message) {
	t := fmt.Sprintf("Привет %s!", msg.From.String())
	sendMessage(msg.Chat.ID, t, msg.MessageID)
	log.Debugf("Say hello to %s", msg.From.String())
}

func commandsHelpHandler(msg *tgbotapi.Message) {
	helpMsg :=
		`Помощь по командам бота.
/start - приветствие (стандартная для любого бота Telegram)
/help - данная справка
/ban @username - забанить пользователя в группе (бот должен иметь административные права в группе)
/unban @username - разбанить пользователя в группе (бот должен иметь административные права в группе)
/ping - шуточный пинг
/yum [info provides repolist repoquery] - аналог системной команды
/dnf [info provides repolist repoquery] - аналог системной команды
`
	sendMessage(msg.Chat.ID, helpMsg, 0)
}

func commandsLinkHandler(msg *tgbotapi.Message) {
	if msg.ReplyToMessage == nil {
		sendMessage(msg.Chat.ID, "Напиши команду в ответ на сообщение, тогда сработает.", msg.MessageID)
		return
	}
	if len(msg.Chat.UserName) == 0 {
		sendMessage(msg.Chat.ID, fmt.Sprintf("Это не публичный чат, ссылку получить невозможно. Message ID = *%d*", msg.ReplyToMessage.MessageID), msg.MessageID)
		return
	}

	sendMessage(msg.Chat.ID, fmt.Sprintf("https://t.me/%s/%d", msg.Chat.UserName, msg.ReplyToMessage.MessageID), msg.MessageID)
}

func commandsPIDHandler(msg *tgbotapi.Message) {
	if msg.ReplyToMessage == nil {
		sendMessage(msg.Chat.ID, "Напиши команду в ответ на сообщение, тогда сработает.", msg.MessageID)
		return
	}

	sendMessage(msg.Chat.ID, fmt.Sprintf("``` %d ```", msg.ReplyToMessage.MessageID), msg.MessageID)
}

func commandsPingHandler(msg *tgbotapi.Message) {
	r := rand.New(rand.NewSource(int64(msg.From.ID)))
	r.Seed(int64(msg.MessageID))

	if r.Int()%12 == 0 {
		sendMessage(msg.Chat.ID, "Request timed out 😜", msg.MessageID)
		return
	}
	pingMsg := fmt.Sprintf("%s пинг от тебя %3.3f 😜", msg.From.String(), r.Float32())
	sendMessage(msg.Chat.ID, pingMsg, msg.MessageID)
}

func commandsFloodHandler(msg *tgbotapi.Message) {
	if !isMeAdmin(msg.Chat) {
		sendMessage(msg.Chat.ID, "Бот не является администратором этого чата. Команда недоступна!", msg.MessageID)
		log.Warn("Command `flood` in chat with bot not admin from %s", msg.From.String())
		return
	}

	if msg.ReplyToMessage == nil {
		sendMessage(msg.Chat.ID, "Напиши команду в ответ на сообщение-флуд, тогда сработает.", msg.MessageID)
		return
	}

	if botUser, err := bot.GetMe(); err != nil {
		log.Errorf("Unable to get bot user: %s", err)
		return
	} else if botUser.ID == msg.ReplyToMessage.From.ID {
		sendMessage(msg.Chat.ID, fmt.Sprintf("Хорошая попытка %s 😜", msg.From.String()), msg.MessageID)
		return
	}

	// chech himself
	if msg.ReplyToMessage.From.ID == msg.From.ID {
		sendMessage(msg.Chat.ID, "Самотык? 😜", msg.MessageID)
		return
	}

	// check flood duration
	if exists, d, err := cacheGet(msg.ReplyToMessage.From.ID, msg.From.ID); err != nil {
		log.Errorf("Unable to get cache: %s")
		return
	} else if exists {
		sendMessage(msg.Chat.ID, fmt.Sprintf("Ты недавно уже объявлял %s флудером. Подожди некоторое время: %s", msg.ReplyToMessage.From.String(), (options.CacheDuration-d).String()), msg.MessageID)
		return
	} else {
		if err = cacheSet(msg.ReplyToMessage.From.ID, msg.From.ID); err != nil {
			log.Errorf("Unable to set cache for flooder ID %d and user ID %d: %s", msg.ReplyToMessage.From.ID, msg.From.ID, err)
		}
	}

	var (
		level   int
		err     error
		apiResp tgbotapi.APIResponse
	)

	if level, err = dbAddFloodLevel(msg.ReplyToMessage.From.ID); err != nil {
		log.Errorf("Unable to add flood level for %d: %s", msg.ReplyToMessage.From.ID, err)
		return
	}
	if level >= options.MaximumFloodLevel {
		config := tgbotapi.ChatMemberConfig{
			ChatID:             msg.Chat.ID,
			SuperGroupUsername: msg.Chat.UserName,
			UserID:             msg.ReplyToMessage.From.ID,
		}
		if apiResp, err = bot.KickChatMember(config); err != nil {
			if apiResp.Ok {
				sendMessage(msg.Chat.ID, fmt.Sprintf("%s терпение туземцев этого чата по поводу твоего флуда кончилось. Мы изгоняем тебя!", msg.ReplyToMessage.From.String()), 0)
			} else {
				log.Warnf("Unable to ban flooder %s. API response with error: (%d) %s", msg.ReplyToMessage.From.String(), apiResp.ErrorCode, apiResp.Description)
			}
		}

		if err = dbSetFloodLevel(msg.ReplyToMessage.From.ID, 0); err != nil {
			log.Errorf("Unable to clear flood level for banned user: %s", err)
		}
	} else {
		sendMessage(msg.Chat.ID, fmt.Sprintf("%s тебя назвали флудером, осталось попыток %d и будешь изгнан!", msg.ReplyToMessage.From.String(), options.MaximumFloodLevel-level), msg.ReplyToMessage.MessageID)
	}
}

func commandsBanHandler(msg *tgbotapi.Message) {
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		sendMessage(msg.Chat.ID, "Кого будем банить в привате? 😂", msg.MessageID)
		log.Debugf("Commands `ban` or `unban` in private chat from %s", msg.From.String())
		return
	}

	if !isMeAdmin(msg.Chat) {
		sendMessage(msg.Chat.ID, "Бот не является администратором этого чата. Команда недоступна!", msg.MessageID)
		log.Warn("Commands `ban` or `unban` in chat with bot not admin from %s", msg.From.String())
		return
	} else {
		log.Debugf("Commands `ban` or `unban` in group or supergroup chat with bot admin from %s", msg.From.String())
	}

	if !isUserAdmin(msg.Chat, msg.From) {
		sendMessage(msg.Chat.ID, "Ты не админ в этом чате! Не имеешь право на баны/разбаны! 🤔\nПопытка управления реальностью записана в аналы, группа немедленного БАНения уже выехала за тобой!😉", msg.MessageID)
		log.Warnf("Commands `ban` or `unban` run fails, user %s not admin in chat!", msg.From.String())
		return
	}

	if msg.CommandArguments() == "" {
		sendMessage(msg.Chat.ID, "Кого будем банить?", msg.MessageID)
		log.Debugf("Command `ban` without arguments from %s", msg.From.String())
		return
	}

	username := msg.CommandArguments()
	var (
		user    *tgbotapi.User
		err     error
		apiResp tgbotapi.APIResponse
	)
	if user, err = getUser(username); err != nil {
		if err == ErrorUserNotFound {
			sendMessage(msg.Chat.ID, fmt.Sprintf("Не нашли пользователя %s", username), msg.MessageID)
			return
		} else if strings.Contains(err.Error(), "Список:") {
			sendMessage(msg.Chat.ID, fmt.Sprintf("Более одного пользователя попало в выборку. Попробуй с @username. \n%s", err), msg.MessageID)
			return
		}
		log.Errorf("Unable to find user with name [%s]: %s", username, err)
		return
	}
	log.Debugf("Found user [%+v]", *user)

	config := tgbotapi.ChatMemberConfig{
		ChatID:             msg.Chat.ID,
		SuperGroupUsername: msg.Chat.UserName,
		UserID:             user.ID,
	}

	if strings.ToLower(msg.Command()) == "ban" {
		apiResp, err = bot.KickChatMember(config)
	} else if strings.ToLower(msg.Command()) == "unban" {
		apiResp, err = bot.UnbanChatMember(config)
	}

	if err != nil {
		if apiResp.Ok || apiResp.ErrorCode == 0 {
			sendMessage(msg.Chat.ID, "Сделано", msg.MessageID)
			log.Debugf("Ban/Unban %s successful", user.String())
		} else {
			sendMessage(msg.Chat.ID, fmt.Sprintf("*Ошибка*: ``` код=%d, описание=%s ```", apiResp.ErrorCode, apiResp.Description), msg.MessageID)
			log.Warnf("API response with error: (%d) %s", apiResp.ErrorCode, apiResp.Description)
		}
	}
}

func commandsDNFHandler(msg *tgbotapi.Message) {
	var (
		err    error
		output []byte
	)
	args := strings.Replace(msg.CommandArguments(), "—", "--", -1)
	if args == "" {
		sendMessage(msg.Chat.ID, "Не знаю, что выполнять, ты же ничего не указал в аргументах", msg.MessageID)
		log.Debugf("Command `dnf` without arguments from %s", msg.From.String())
		return
	}

	arglist := strings.Split(args, " ")
	if arglist[0] == "info" || arglist[0] == "provides" || arglist[0] == "repolist" || arglist[0] == "repoquery" {
		if arglist[0] != "repolist" && arglist[0] != "repoquery" {
			arglist = append(arglist, "-q")
		}
		cmd := exec.Command("/usr/bin/dnf", arglist...)
		/*var (
			stdout io.ReadCloser
			stderr io.ReadCloser
		)
		if stdout, err = cmd.StdoutPipe(); err != nil {
			log.Errorf("Unable to get stdout pipe: %s", err)
			return
		}
		if stderr, err = cmd.StderrPipe(); err != nil {
			log.Errorf("Unable to get stderr pipe: %s", err)
			return
		}

		if err = cmd.Start(); err != nil {
			log.Errorf("Unable to start command [dnf %s]: %s", strings.Join(arglist, " "), err)
			return
		}

		var buf []byte
		if _, err = stdout.Read(buf); err != nil {
			log.Errorf("Unable to read stdout for command [dnf %s]: %s", strings.Join(arglist, " "), err)
			return
		}
		if len(buf) > 0 {
			output = append(output, buf...)
		}
		if _, err = stderr.Read(buf); err != nil {
			log.Errorf("Unable to read stderr for command [dnf %s]: %s", strings.Join(arglist, " "), err)
			return
		}
		if len(buf) > 0 {
			output = append(output, buf...)
		}

		if err = cmd.Wait(); err != nil {
			log.Errorf("Unable to wait command [dnf %s]: %s", strings.Join(arglist, " "), err)
			return
		}

		if len(output) > 0 {
			sendMessage(msg.Chat.ID, fmt.Sprintf("``` %s ```", output), msg.MessageID)
			log.Debugf("Run command from %s: dnf %s", msg.From.String(), strings.Join(arglist, " "))
		} else {
			sendMessage(msg.Chat.ID, "А нечего выводить, вывод пустой", msg.MessageID)
			log.Warnf("Run command from %s: dnf %s with empty output", msg.From.String(), strings.Join(arglist, " "))
		}*/

		if output, err = cmd.CombinedOutput(); err != nil {
			log.Errorf("Unable to run command form %s: dnf %s: %s", msg.From.String(), strings.Join(arglist, " "), strings.Join(arglist, " "))
			sendMessage(msg.Chat.ID, "Ой. Что-то пошло не так!", msg.MessageID)
		} else if len(output) == 0 {
			sendMessage(msg.Chat.ID, "А нечего выводить, вывод пустой", msg.MessageID)
			log.Warnf("Run command from %s: dnf %s with empty output", msg.From.String(), strings.Join(arglist, " "))
		} else {
			sendMessage(msg.Chat.ID, fmt.Sprintf("``` %s ```", output), msg.MessageID)
			log.Debugf("Run command from %s: dnf %s", msg.From.String(), strings.Join(arglist, " "))
		}
	}
}
