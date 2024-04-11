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
    "fmt"
    "log"
    "time"
    "errors"
    "syscall"
    "net/http"
    "encoding/json"
    "container/list"
)

type WebErrorMessage struct {
    Code     int     `json: error_code`
    Error    string  `json: error`
}

var people *UsersDb

type Client struct {
    id      string
    realip  string
    queue   list.List
}

var clients map[string]*Client


func fatal_error(w http.ResponseWriter, r *http.Request, e error, sent bool) {

    var errmsg WebErrorMessage

    if sent == false {
        w.WriteHeader(http.StatusInternalServerError)

        errmsg.Code = 502
        errmsg.Error = e.Error()

        txt, err := json.Marshal(errmsg)
        if err != nil {
            return
        }

        w.Write([]byte(txt))
    }

    log.Printf("fatal error while handling request: %v", e)
}


func request_handler(w http.ResponseWriter, r *http.Request) {

    var err error
    sent := false

    switch r.Method {

    case "POST":

        switch r.URL.Path {
        case "/updatepos":
            err = handle_position_update(w, r)

        case "/updatestatus":
            err = handle_status_update(w, r)

        default:
            err = errors.New("unsupported endpoint requested")
        }

    case "GET":

        switch r.URL.Path {
        case "/bootstrap":
            err, sent = bootstrap(w, r)

        case "/people":

            var cln *Client

            cln = new(Client)

            cln.realip = r.Header.Get("X-Forwarded-For")
            cln.id = r.RemoteAddr
            clients[cln.id] = cln

            log.Printf("Client %v connected", r.RemoteAddr)

            err, sent = people_event_source(w, r, cln)

            delete(clients, r.RemoteAddr)
            log.Printf("Client: %v done", r.RemoteAddr)

        default:
            err = errors.New("unsupported endpoint requested")
        }

    default:
        err = errors.New("unsupported method")
    }

    if (err != nil) {
        if errors.Is(err, syscall.EPIPE) {
            /* connection closed */
            return
        }

        fatal_error(w, r, err, sent)
        return
    }
}


func handle_position_update(w http.ResponseWriter, r *http.Request) error {

    decoder := json.NewDecoder(r.Body)

    var ui *UserInfo
    var up UserPosition

    err := decoder.Decode(&up)
    if err != nil {
        return err
    }

    ui = people.get(up.UserName, true)
    if ui == nil {
        return fmt.Errorf("failed to get user %v", up.UserName)
    }

    ui.UpdatePosition(&up)

    log.Printf("position update for %s: [lat:%2f, lon:%2f]\n",
               up.UserName, up.Lat, up.Lon)

    for _, client := range clients {
        client.queue.PushBack(ui)
    }

    return nil
}


func handle_status_update(w http.ResponseWriter, r *http.Request) error {

    decoder := json.NewDecoder(r.Body)

    var ui *UserInfo
    var us UserStatus

    err := decoder.Decode(&us)
    if err != nil {
        return err
    }

    ui = people.get(us.UserName, true)
    if ui == nil {
        return fmt.Errorf("failed to get user %v", us.UserName)
    }

    ui.UpdateStatus(&us)

    log.Printf("status update for %s: '%s'\n", us.UserName, us.Status)

    for _, client := range clients {
        client.queue.PushBack(ui)
    }

    return nil
}

func bootstrap(w http.ResponseWriter, r *http.Request) (error, bool) {

    w.Header().Set("Content-Type", "application/json");
    w.Header().Set("Cache-Control", "no-cache");

    txt, err := people.exportJSON()

    _, err = w.Write(txt)
    if (err != nil) {
        return err, false
    }

    if f, ok := w.(http.Flusher); ok {
        f.Flush()
    }

    log.Printf("bootstrap: export done")

    return nil, true
}

func send_event(w http.ResponseWriter, txt string, headers_sent *bool) (error) {

    var err error

    msg := fmt.Sprintf("event: posupdate\ndata: %s\n\n", txt)

    _, err = w.Write([]byte(msg))
    if (err != nil) {
        return err
    }

    if f, ok := w.(http.Flusher); ok {
        f.Flush()
    }

    *headers_sent = true

    return nil
}


func people_event_source(w http.ResponseWriter, r *http.Request, client *Client) (error, bool) {

    w.Header().Set("Content-Type", "text/event-stream");
    w.Header().Set("Cache-Control", "no-cache");

    /*
     * To avoid closing keepalive connection without activity,
     * send something once in interval
     */
    var quiet = 0
    const interval = 30

    var headers_sent = false

    log.Printf("entering event source loop...")

    for {
        qlen := client.queue.Len()

        if (qlen == 0) {
            time.Sleep(1 * time.Second)
            quiet += 1

            if quiet < interval {
                continue
            }

            /* we have no real updates, send small keepalive message */

            var ka KeepalivePing
            ka.Alive = true

            var txt []byte
            var err error

            txt, err = json.Marshal(ka)

            err = send_event(w, string(txt), &headers_sent)
            if err != nil {
                return err, headers_sent
            }

            quiet = 0
            continue
        }

        var out = make([]UserInfo, qlen)

        var i = 0
        for qlen > 0 {
            elem := client.queue.Front()

            v := elem.Value.(*UserInfo)
            out[i] = *v
            i += 1

            client.queue.Remove(elem)
            qlen -= 1
        }

        var txt []byte
        var err error

        txt, err = json.Marshal(out)
        if err != nil {
            return err, headers_sent
        }

        // TODO: check if connection closed by client
        err = send_event(w, string(txt), &headers_sent)
        if err != nil {
            return err, headers_sent
        }

        quiet = 0

        log.Printf("pushed events to client %s(%s): %s, total clients: %d",
                   client.id, client.realip, string(txt), len(clients))

        time.Sleep(1 * time.Second)
    }

    return nil, headers_sent
}

func logRequest(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s[%s] %s '%s'\n",
                    r.Header.Get("X-Forwarded-For"),
                    r.RemoteAddr,
                    r.Method,
                    r.URL)
        handler.ServeHTTP(w, r)
    })
}




func main() {

    if len(os.Args) < 2 {
        log.Printf("Usage: " + os.Args[0] + ": <conf.json>\n" );
        os.Exit(1);
    }

    conf, err := ConfigLoad(os.Args[1])
    if err != nil {
        log.Println("config load failed: " + err.Error())
        os.Exit(1)
    }

    var dcfg DaemonConfig

    dcfg.AppID = "webmap"
    dcfg.LogFile = conf.WebmapLog
    dcfg.Syslog = conf.Syslog
    dcfg.Stderr = conf.Stderr

    err = init_daemon(&dcfg)
    if err != nil {
        log.Println("daemon init failed: " + err.Error())
        os.Exit(1)
    }

    /* webmap only reads state file, shared with bot */
    people, err = CreateUsersDb(conf.StateFile)
    if err != nil {
        log.Println("failed to load users: " + err.Error())
        os.Exit(1)
    }

    clients = make(map[string]*Client)


    log.Printf("webmap server is listening at %s", conf.WebmapListen)

    http.HandleFunc("/", request_handler)
    log.Fatal(http.ListenAndServe(conf.WebmapListen, logRequest(http.DefaultServeMux)))
}
