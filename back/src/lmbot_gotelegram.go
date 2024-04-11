/*
 * Copyright (C) 2024 Vladimir Homutov
 */

/*
 * This file is part of Rieman.
 *
 * Rieman is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Rieman is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 */

package main

import (
    "os"
    "os/signal"
    "context"

    "log"
    "encoding/json"

    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

type LMBot struct {
    ctx            context.Context
    bot           *bot.Bot
    conf          *UserConfig
}

type LMMessage struct {
    Userid         string
    Text           string
    Edited         bool
    Location      *GeoPos
    Status         string

    menu_title     string

    ChatID         int64
    MessageID      int
}

type LMMessageHandler func(bot *LMBot, msg *LMMessage) (error)


func lm_bot_new(conf *UserConfig) (*LMBot, error) {
    var res LMBot

    ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

    res.ctx = ctx
    res.conf = conf


    return &res, nil
}

func bot_msg_handler(ctx context.Context, b* bot.Bot, update *models.Update,
                     lmbot *LMBot, handler LMMessageHandler) {

    var lm_msg   LMMessage
    var msg     *models.Message

    if update.EditedMessage != nil {
        msg = update.EditedMessage
        lm_msg.Edited = true

    } else if update.Message != nil {
        msg = update.Message
        lm_msg.Edited = false

    } else {
        return
    }

    lm_msg.ChatID = msg.Chat.ID
    lm_msg.MessageID = msg.ID

    //debug_input_msg(msg)

    lm_msg.Userid = msg.From.FirstName

    if user_allowed(ctx, b, msg, lmbot.conf.RestrictChannelId) == false {
        //lm_bot_reply_to(lmbot, &lm_msg, "you are not allowed")

        /* ignore messages from unauthorized persons */
        return
    }

    if msg.Location == nil {
        lm_msg.Location = nil

    } else {
        var pos GeoPos
        pos.Lon = msg.Location.Longitude
        pos.Lat = msg.Location.Latitude

        lm_msg.Location = &pos
    }

    lm_msg.Text = msg.Text

    err := handler(lmbot, &lm_msg)
    if err != nil {
        log.Printf("oops shit");
    }
}


func lm_bot_process_messages(lmbot *LMBot, handler LMMessageHandler) (error) {

    opts := []bot.Option{
        bot.WithDefaultHandler(
            func (ctx context.Context, b* bot.Bot, update *models.Update) {
                  bot_msg_handler(ctx, b, update, lmbot, handler);
            }),
        bot.WithCallbackQueryDataHandler("status_", bot.MatchTypePrefix,
            func (ctx context.Context, b *bot.Bot, update *models.Update) {
                  bot_menu_handler(ctx, b, update, lmbot, handler);
            }),
    }

    bot, err := bot.New(lmbot.conf.Token, opts...)
    if err != nil {
        return err
    }

    //log.Printf("Authorized on account %s", bot.Self.UserName)

    lmbot.bot = bot

    // pass handler now to custom function
    lmbot.bot.Start(lmbot.ctx)

    log.Printf("started bot");

    return nil
}

func bot_menu_handler(ctx context.Context, b* bot.Bot, update *models.Update,
                      lmbot *LMBot, handler LMMessageHandler) {

    var lm_msg   LMMessage

    log.Printf("menu handler: clicked");

    b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
        CallbackQueryID: update.CallbackQuery.ID,
        ShowAlert:       false,
    })

    lm_msg.Userid = update.CallbackQuery.From.FirstName
    lm_msg.ChatID = update.CallbackQuery.Message.Chat.ID
    lm_msg.MessageID = update.CallbackQuery.Message.MessageID
    lm_msg.Status = update.CallbackQuery.Data

    log.Printf("menu handler: got status: %v", lm_msg.Status);

    handler(lmbot, &lm_msg);
}

func debug_input_msg(m *models.Message) {
    dump, _ := json.Marshal(m)
    log.Printf("MSG=%s\n", dump)
    log.Printf("\n\n\n")
}

func lm_bot_react(lmbot *LMBot, lm_msg *LMMessage, emoji string) {

    var react bot.SetMessageReactionParams

    var pbool = false

    react.ChatID = lm_msg.ChatID
    react.MessageID = lm_msg.MessageID
    react.IsBig = &pbool

    var rtypes []models.ReactionType

    var rt models.ReactionType

    var re models.ReactionTypeEmoji

    re.Type = "emoji"
    re.Emoji = emoji

    rt.ReactionTypeEmoji = &re

    rtypes = append(rtypes, rt)

    react.Reaction = rtypes

    _, err := lmbot.bot.SetMessageReaction(lmbot.ctx, &react)
    if err != nil {
        log.Printf("failed to react: %v", err);
    }
}


func get_statuses() map[string]string {

    var statuses = map[string]string{
        STATUS_MOVING:  i18n[STR_STATUS_MOVING],
        STATUS_PITSTOP: i18n[STR_STATUS_PITSTOP],
        STATUS_PUNCTURE: i18n[STR_STATUS_PUNCTURE],
        STATUS_FALL: i18n[STR_STATUS_FALL],
        STATUS_INCIDENT: i18n[STR_STATUS_INCIDENT],
        STATUS_FINISHED: i18n[STR_STATUS_FINISHED],
        STATUS_DNF: i18n[STR_STATUS_DNF],
    }

    return statuses
}

func menu_title(status string, current string) string {

    statuses := get_statuses()

    if (status == current) {
        return " ** " + statuses[status] + " ** "
    } else {
        return statuses[status]
    }
}

func lm_bot_send_menu(lmbot *LMBot, lm_msg *LMMessage) {

    s := lm_msg.Status

    kb := &models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{
            {
                {Text: menu_title(STATUS_MOVING, s), CallbackData: STATUS_MOVING},
                {Text: menu_title(STATUS_PITSTOP, s), CallbackData: STATUS_PITSTOP},
            }, {
                {Text: menu_title(STATUS_PUNCTURE, s) , CallbackData: STATUS_PUNCTURE},
                {Text: menu_title(STATUS_FALL, s), CallbackData: STATUS_FALL},
                {Text: menu_title(STATUS_INCIDENT, s), CallbackData: STATUS_INCIDENT},
            },
            {
                {Text: menu_title(STATUS_FINISHED, s), CallbackData: STATUS_FINISHED},
                {Text: menu_title(STATUS_DNF, s), CallbackData: STATUS_DNF},
            },
        },
    }

    var shortcut = `| /start | <a href="` + lmbot.conf.LiveMapURL + `">`+i18n[STR_LIVE_MAP]+`</a> |`

    var msg    bot.SendMessageParams

    msg.ChatID = lm_msg.ChatID
    msg.ReplyMarkup = kb
    msg.Text = lm_msg.menu_title + "          " + shortcut
    msg.ParseMode = models.ParseModeHTML


    _, err := lmbot.bot.SendMessage(lmbot.ctx, &msg)
    if (err != nil) {
        log.Printf("failed to send menu: %v", err);
    }
}


func lm_bot_reply_to(lmbot *LMBot, lm_msg *LMMessage, reply_text string) {

    var msg    bot.SendMessageParams
    var reply  models.ReplyParameters

    msg.ChatID = lm_msg.ChatID
    msg.Text = reply_text

    reply.MessageID = lm_msg.MessageID
    msg.ReplyParameters = &reply

    _, err := lmbot.bot.SendMessage(lmbot.ctx, &msg)
    if (err != nil) {
        log.Printf("failed to reply: %v", err);
    }
}

func lmbot_send_msg(lmbot *LMBot, lm_msg *LMMessage, text string, html bool) {

    var msg    bot.SendMessageParams

    msg.ChatID = lm_msg.ChatID
    msg.Text = text

    if (html) {
        msg.ParseMode = models.ParseModeHTML
    }

    _, err := lmbot.bot.SendMessage(lmbot.ctx, &msg)
    if (err != nil) {
        log.Printf("failed to send message: %v", err);
    }
}

func user_allowed(ctx context.Context, b *bot.Bot, m *models.Message,
                  channelid int64) bool {

    var cf bot.GetChatMemberParams
    var cm *models.ChatMember

    cf.ChatID = channelid
    cf.UserID = m.From.ID

    cm, err := b.GetChatMember(ctx, &cf)
    if (err != nil) {
        log.Printf("bot.GetChatMember() error: %v", err)
        return false
    }

    //log.Printf("got chat membership info: %v", cm)

    if cm.Left == nil {
        return true
    }

    return cm.Left.Status != "left"
}
