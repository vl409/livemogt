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
    "log"
    "fmt"
    "bytes"
    "time"
    "net/http"
    "encoding/json"
)

var i18n map[int]string;

// https://core.telegram.org/bots/api#reactiontype
const REACT_OK = "👌"

var people *UsersDb

func handle_status_update(conf *UserConfig, up UserStatus) (error) {

    j, err := json.Marshal(up)
    if (err != nil) {
        return fmt.Errorf("JSON creation failed: %v", err)
    }

    err = send_update(conf.UpdateStatusURL, j)
    if (err != nil) {
        return fmt.Errorf("failed to send status: %v", err)
    }

    return nil
}


func handle_position_update(conf *UserConfig, up UserPosition) (error) {

    j, err := json.Marshal(up)
    if (err != nil) {
        return fmt.Errorf("JSON creation failed: %v", err)
    }

    err = send_update(conf.UpdatePositionURL, j)
    if (err != nil) {
        return fmt.Errorf("failed to send position: %v", err)
    }

    return nil
}

func send_update(url string, data []byte) error {

    req, err := http.NewRequest("POST", url, bytes.NewReader(data))
    if (err != nil) {
        return err
    }

    req.Header.Set("Content-Type", "application/json")

    client := http.Client{Timeout: 10 * time.Second}
    res, err := client.Do(req)
    if err != nil {
        return err
    }

    log.Printf("%s => '%s': %d\n", data, url, res.StatusCode)

    return nil
}

func create_menu_header(userid string, status string) string {
    if len(status) == 0 {
        return "<b>" + userid + "</b> "
    }
    return "<b>" + userid + "</b> (" + status + ")"
}


func handle_message(bot *LMBot, msg *LMMessage) (error) {

    var user *UserInfo

    user = people.get(msg.Userid, false)

    if (msg.Text == "/start") {
        var start_text = i18n[STR_HTML_START]
        lmbot_send_msg(bot, msg, start_text, true)
        return nil
    }

    if (user == nil) {
        /* new user - perform some introduction */

        var s string

        if (msg.Location == nil) {
            s = fmt.Sprintf(i18n[STR_FMT_GEO_REQUEST], msg.Userid)
            if !msg.Edited {
                lm_bot_reply_to(bot, msg, s)
                log.Printf("greeted unknown user %s", msg.Userid)
            }
            return nil
        }

        var up UserPosition

        up.UserName = msg.Userid
        up.Lat = msg.Location.Lat
        up.Lon = msg.Location.Lon
        up.Last = time.Now()

        user = createUser(nil, &up) /* always ok */
        people.set(msg.Userid, user)

        if !msg.Edited {
            if (len(msg.Status) == 0) {
                msg.Status = user.MovingState
            }

            msg.menu_title = create_menu_header(msg.Userid, user.Status)
            s = fmt.Sprintf(i18n[STR_FMT_WELCOME_GOT_GEO], msg.Userid)
            lm_bot_reply_to(bot, msg, s)
            lm_bot_send_menu(bot, msg)
        }

        log.Printf("created new user %s", up.UserName)
    }

    if (msg.Text == "/status" || len(msg.Status) != 0) {

        if (len(msg.Status) == 0) {
            msg.Status = user.MovingState

        } else {
            var up UserStatus

            up.UserName = msg.Userid
            up.MovingState = msg.Status

            user.UpdateStatus(&up)

            err := handle_status_update(bot.conf, up)
            if err != nil {
                log.Printf("error while sending status update: %v", err)
            }
        }

        msg.menu_title = create_menu_header(msg.Userid, user.Status)
        lm_bot_send_menu(bot, msg)

    } else if (msg.Location != nil) {

        var up UserPosition

        up.UserName = msg.Userid
        up.Lat = msg.Location.Lat
        up.Lon = msg.Location.Lon
        up.Last = time.Now()

        user.UpdatePosition(&up)

        err := handle_position_update(bot.conf, up)
        if err != nil {
            log.Printf("error while sending position update: %v", err)
        }

        if !msg.Edited {
            lm_bot_react(bot, msg, REACT_OK)
        }

    } else if (len(msg.Text) != 0) {

        if len(msg.Text) > bot.conf.MaxStatus {
            log.Printf("Status too long")
            lm_bot_reply_to(bot, msg, i18n[STR_STATUS_TOO_LONG])
            return nil
        }

        var up UserStatus

        up.UserName = msg.Userid
        up.Status = msg.Text

        user.UpdateStatus(&up)

        err := handle_status_update(bot.conf, up)
        if err != nil {
            log.Printf("error while sending status update: %v", err)
        }

        if !msg.Edited {

            if (len(msg.Status) == 0) {
                msg.Status = user.MovingState
            }

            //lm_bot_react(bot, msg, REACT_OK)
            msg.menu_title = create_menu_header(msg.Userid, user.Status)
            lm_bot_send_menu(bot, msg)
        }
    }

    log.Printf("done with message [edit:%v], total users: %d", msg.Edited, people.count())

    people.set(msg.Userid, user)

    err := people.save(bot.conf.TmpDir)
    if err != nil {
        log.Printf("failed to update state file: %v", err)
    }

    return nil
}


func main() {

    if len(os.Args) < 2 {
        log.Printf("Usage: " + os.Args[0] + ": <conf.json>\n" );
        os.Exit(1);
    }

    conf, err := ConfigLoad(os.Args[1])
    if err != nil {
        log.Println(err.Error())
        os.Exit(1)
    }

    i18n, err = get_i18n(&conf)
    if err != nil {
        log.Println(err.Error())
        os.Exit(1)
    }

    var dcfg DaemonConfig

    dcfg.AppID = "livemogt"
    dcfg.LogFile = conf.BotLog
    dcfg.Syslog = conf.Syslog
    dcfg.Stderr = conf.Stderr

    err = init_daemon(&dcfg)
    if err != nil {
        log.Println("daemon init failed: " + err.Error())
        os.Exit(1)
    }

    people, err = CreateUsersDb(conf.StateFile)
    if err != nil {
        log.Println("failed to init users db: " + err.Error())
        os.Exit(1)
    }

    bot, err := lm_bot_new(&conf)
    if err != nil {
        log.Println(err.Error())
        os.Exit(1)
    }

    err = lm_bot_process_messages(bot, handle_message)
    if err != nil {
        log.Println(err.Error())
        os.Exit(1)
    }
}
