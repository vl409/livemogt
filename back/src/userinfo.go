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
    "time"
    "encoding/json"
)

const TrackDepth = 64

/* must start with same prefix to match in filter */
const STATUS_MOVING = "status_moving"
const STATUS_PITSTOP = "status_pitstop"
const STATUS_PUNCTURE = "status_puncture"
const STATUS_FALL = "status_fall"
const STATUS_INCIDENT = "status_incident"
const STATUS_FINISHED = "status_finished"
const STATUS_DNF = "status_dnf"

type UserMap = map[string]*UserInfo

/* map with users, serialized to/from StateFile */
type UsersDb struct {
    people     UserMap
    StateFile  string
}


type GeoPos struct {
    Lon      float64
    Lat      float64
}

/* all information we know about user */
type UserInfo struct {
    UserName     string
    Status       string
    MovingState  string
    Pos          GeoPos
    Last         time.Time
    Track       *RingBuffer
}


func CreateUsersDb(statefile string) (*UsersDb, error) {

    db := new(UsersDb)

    db.people = make(UserMap)
    db.StateFile = statefile

    err := db.load()
    if err != nil {
        return nil, err
    }

    return db, nil
}

func (db *UsersDb) get(name string, create bool) (*UserInfo) {

    var ok  bool
    var ui *UserInfo

    ui, ok = db.people[name]

    if (!ok) {
        if create == true {
            ui = createUser(nil, nil)
            if ui == nil {
                return nil
            }

            db.set(name, ui)

        } else {
            return nil
        }
    }

    return ui
}

func (db *UsersDb) set(userid string, ui *UserInfo) {
    ui.UserName = userid
    db.people[userid] = ui
}

func (db *UsersDb) count() int {
    return len(db.people)
}


func (db *UsersDb) load() error {

    loadfrom := db.StateFile

    _, err := os.Stat(loadfrom)
    if os.IsNotExist(err) {
        log.Printf("state file not found, ignored")
        return nil
    }

    f, err := os.ReadFile(loadfrom)
    if err != nil {
        log.Printf("failed to read file '%s': %v", loadfrom, err)
        return err
    }

    var users []UserInfo

    err = json.Unmarshal(f, &users)
    if err != nil {
        log.Printf("failed to parse file '%s': %v", loadfrom, err)
        return err
    }

    var i = 0

    for _, v := range users {

        ui := createUser(nil, nil);

        ui.UserName = v.UserName
        ui.Status = v.Status

        if len(v.MovingState) == 0 {
            ui.MovingState = STATUS_MOVING
        } else {
            ui.MovingState = v.MovingState
        }

        ui.Pos = v.Pos
        ui.Last = v.Last
        ui.Track = v.Track

        db.set(v.UserName, ui)
        log.Printf("loaded user '%v' from state", v.UserName)
        i += 1
    }

    log.Printf("state file loaded, %d users found", i)

    return nil
}


func (db *UsersDb) save(tmpdir string) error {

    txt, err := db.exportJSON()
    if err != nil {
        return fmt.Errorf("failed to export JSON: %v", err)
    }

    f, err := os.CreateTemp(tmpdir, "")
    if err != nil {
        return fmt.Errorf("failed to open temp file: %v", err)
    }

    _, err = f.Write([]byte(txt))
    if err != nil {
        os.Remove(f.Name())
        return err
    }

    err = os.Rename(f.Name(), db.StateFile)

    if err != nil {
        os.Remove(f.Name())
        return err
    }

    return nil
}


func (db *UsersDb) exportJSON() ([]byte, error) {

    var out = make([]UserInfo, people.count())

    var i = 0

    /* convert map to array for serializing */
    for _, v := range people.people {
        out[i] = *v
        i += 1
    }

    log.Printf("export: %d user(s) serialized", i)

    return json.Marshal(out)
}

func createUser(us *UserStatus, up *UserPosition) *UserInfo{

    ui := new(UserInfo)

    ui.Track = CreateRing(TrackDepth)

    if us != nil {
        ui.Status = us.Status
        ui.MovingState = us.MovingState
    }

    if up != nil {
        ui.Pos.Lat = up.Lat
        ui.Pos.Lon = up.Lon
        ui.Last = up.Last
    }

    return ui
}

func (ui *UserInfo) UpdatePosition(up *UserPosition) {

    zeroed := (ui.Pos.Lat == 0 && ui.Pos.Lon == 0)
    changed := (up.Lat != ui.Pos.Lat || up.Lon != ui.Pos.Lon)

    /* avoid pushing initial and current states to track */
    if (changed && !zeroed) {
        ui.Track.push(ui.Pos)
    }

    ui.Pos.Lat = up.Lat
    ui.Pos.Lon = up.Lon
    ui.Last = up.Last

    log.Printf("updated position for user %s", ui.UserName)
}

func (ui *UserInfo) UpdateStatus(us *UserStatus) {
    if (len(us.Status) != 0) {
        ui.Status = us.Status
        log.Printf("updated status for user %s", ui.UserName)
    }

    if (len(us.MovingState) != 0) {
        ui.MovingState = us.MovingState
        log.Printf("updated moving state for user %s", ui.UserName)
    }

}


func (u *UserInfo) MarshalJSON() ([]byte, error) {

    type Alias UserInfo

    positions := u.Track.extract()

    aux := &struct {
        Track []GeoPos `json:"Track,omitempty"`
        *Alias
    } {
        Track: positions,
        Alias: (*Alias)(u),
    }

    return json.Marshal(aux)
}


func (ui *UserInfo) UnmarshalJSON(data []byte) error {

    type Alias UserInfo

    aux := &struct {
        Track []GeoPos `json:"Track"`
        *Alias
    } {
        Alias: (*Alias)(ui),
    }

    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }

    ui.Track = CreateRing(TrackDepth)
    for _, pos := range(aux.Track) {
        ui.Track.push(pos)
    }

    return nil
}



