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
    "fmt"
    "os"
    "time"
    "bytes"
    "net/http"
    "math/rand"
    "encoding/json"
    "github.com/tkrajina/gpxgo/gpx"
    rn "github.com/random-names/go"
)

type FakeUser struct {
    index      int
    name       string
    pos        UserPosition
    status     UserStatus
    next_move  time.Time
}

type SimplePoint struct {
    Lat      float64
    Lon      float64
}

const NUsers = 20

var NPoints = 0


var users [NUsers]FakeUser
var points []SimplePoint

func getNames() {

    db := "census-90/male.first"
    fns, err := rn.GetRandomNames(db, &rn.Options{Number:NUsers})
    if err != nil {
        fmt.Printf("failed to get %v names: from %v: %v ",
                   NUsers, db, err.Error())
        os.Exit(1)
    }

    db = "census-90/all.last"
    lns, err := rn.GetRandomNames(db, &rn.Options{Number:NUsers})
    if err != nil {
        fmt.Printf("failed to get %v names: from %v: %v ",
                   NUsers, db, err.Error())
        os.Exit(1)
    }

    for i := 0; i < NUsers; i++ {
        users[i].name = fns[i] + " " + lns[i]
        users[i].next_move = time.Now()

        fmt.Printf("%s\n", users[i].name)
    }
}

func get_rand(min int, max int) int {

    return rand.Intn(max - min + 1) + min

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

    fmt.Printf("%s => '%s': %d\n", data, url, res.StatusCode)

    return nil
}

var statuses = []string {
    "status_moving",
    "status_pitstop",
    "status_fall",
    "status_finished",
    "status_dnf",
}

func move_user(u *FakeUser) {

    step := get_rand(3, 10)

    u.index = (u.index + step) % NPoints

    point := points[u.index]

    var delta float64

    var off = get_rand(-1, 1)
    delta = float64(off) / 1000.0

    var slen = len(statuses)
    var sindex = 0

    // make special status a bit more rare
    for i := 0; i < 3; i++ {
        sindex = get_rand(0, slen - 1)
        if statuses[sindex] == "status_fall" {
            continue
        }
    }

    u.pos.UserName = u.name
    u.pos.Lat = point.Lat + delta
    u.pos.Lon = point.Lon + delta
    u.pos.Last = time.Now()

    u.status.UserName = u.name
    u.status.MovingState = statuses[sindex]

    delay := time.Duration(get_rand(1, 20)) * time.Second

    u.next_move = u.next_move.Add(delay)

    fmt.Printf("move user: %s step: %d idx=%d pos: [%f,%f] next in: %v\n",
               u.name, step, u.index, point.Lat, point.Lon, delay);
}


func main() {

    getNames()

    bytes, err := os.ReadFile("./track.gpx")
    if err != nil {
        os.Stderr.WriteString("failed to load GPX: " + err.Error() + "\n")
        os.Exit(1)
    }

    gpxFile, err := gpx.ParseBytes(bytes)
    if err != nil {
        os.Stderr.WriteString("failed to parse GPX: " + err.Error() + "\n")
        os.Exit(1)
    }

    track := gpxFile.Tracks[0]

    for _, segment := range track.Segments {
        for _, point := range segment.Points {
            var p SimplePoint

            p.Lat = point.Point.Latitude
            p.Lon = point.Point.Longitude

            points = append(points, p)
        }
    }

    NPoints = len(points)

    for {
        time.Sleep(1 * time.Second)

        var now = time.Now()

        for i := 0; i < NUsers; i++ {

            if (users[i].next_move.After(now)) {
                continue
            }

            move_user(&users[i])

            txt, err := json.Marshal(users[i].pos)
            if err != nil {
                fmt.Printf("failed to encode user position: %v\n", err)
                os.Exit(1)
           }

            err = send_update("http://127.0.0.1:8234/updatepos", txt)
            if (err != nil) {
                fmt.Printf("failed to send position: %v", err)
                os.Exit(1)
            }

            txt, err = json.Marshal(users[i].status)
            if err != nil {
                fmt.Printf("failed to encode user status: %v\n", err)
                os.Exit(1)
           }

            err = send_update("http://127.0.0.1:8234/updatestatus", txt)
            if (err != nil) {
                fmt.Printf("failed to send status: %v", err)
                os.Exit(1)
            }
        }

    }
}
