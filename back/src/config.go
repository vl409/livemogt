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
    "encoding/json"
)

type UserConfig struct {
    Token             string
    WebmapListen      string
    WebmapLog         string
    BotLog            string
    BotLang           string
    Syslog            bool
    Stderr            bool
    UpdatePositionURL string
    UpdateStatusURL   string
    LiveMapURL        string
    MaxStatus         int
    StateFile         string
    RestrictChannelId int64
    TmpDir            string
}


func ConfigLoad(fn string) (UserConfig, error) {
    var conf UserConfig

    bytes, err := os.ReadFile(fn)
    if err != nil {
        return conf, err
    }

    err = json.Unmarshal(bytes, &conf)
    if err != nil {
        return conf, err
    }

    return conf, nil
}

