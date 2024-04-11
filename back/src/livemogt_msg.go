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
)

const  (
    STR_HTML_START = iota
    STR_FMT_GEO_REQUEST
    STR_FMT_WELCOME_GOT_GEO
    STR_STATUS_TOO_LONG
    STR_USER_NOT_ALLOWED
    STR_STATUS_MOVING
    STR_STATUS_PITSTOP
    STR_STATUS_PUNCTURE
    STR_STATUS_FALL
    STR_STATUS_INCIDENT
    STR_STATUS_FINISHED
    STR_STATUS_DNF
    STR_LIVE_MAP
)

func get_i18n(conf *UserConfig) (map[int]string, error) {

    if conf.BotLang != "en" && conf.BotLang != "ru" {
        return nil, fmt.Errorf("unsupported language: %v", conf.BotLang)
    }

    var data = map[string] map[int]string {
    "en": {
        STR_HTML_START: `<b>Welcome to LiveMOGT bot</b>
* Type /start to repeat this message
* Share your position with the bot (Attach->Geo->Translate my Position)
* Any text message to will update your profile info (whatever you like to share: phone, email, real name...)
* Type /status to set your status via menu
* Visit <a href="` + conf.LiveMapURL + `">Live map</a> that tracks everyone!`,

        STR_FMT_GEO_REQUEST: `Hello, %s. Translate me your Live GEO position to start`,
        STR_FMT_WELCOME_GOT_GEO: `Welcome, %s. Got your GEO, check further actions in menu`,
        STR_STATUS_TOO_LONG: `status ignore - too long`,
        STR_USER_NOT_ALLOWED: `you are not allowed`,
        STR_STATUS_MOVING: `OK, moving`,
        STR_STATUS_PITSTOP: `Pit-Stop`,
        STR_STATUS_PUNCTURE: `Puncture`,
        STR_STATUS_FALL: `Fall`,
        STR_STATUS_INCIDENT: `Road Incident`,
        STR_STATUS_FINISHED: `Finished`,
        STR_STATUS_DNF: `DNF`,
        STR_LIVE_MAP: `Live map`,
    },

    "ru": {
        STR_HTML_START: `<b>Добро пожаловать в LiveMOGT!</b>
* Отправьте /start чтобы увидеть это сообщение снова
* Поделитесь с ботом свей геопозицей (Attach->Geo->Translate my Position)
* Любое текстовое сообщение боту обновит ваш профиль (что угодно, чем хотите поделиться: почта, телефон, имя...)
* Отправьте /status чтобы увидеть меню и управлять вашим статусом
* Отслеживайте всех на <a href="` + conf.LiveMapURL + `">интерактивной карте</a>!`,

        STR_FMT_GEO_REQUEST: `Привет, %s. Начните трансляцию своей геопозиции, чтобы начать работу с ботом`,
        STR_FMT_WELCOME_GOT_GEO: `Добро пожаловать, %s. Ваша позиция полученая, управляйте всем из меню`,
        STR_STATUS_TOO_LONG: `слишком длинный статус - проигнорирован`,
        STR_USER_NOT_ALLOWED: `вам запрещён доступ к боту`,
        STR_STATUS_MOVING: `OK, еду`,
        STR_STATUS_PITSTOP: `Плановая остановка`,
        STR_STATUS_PUNCTURE: `Прокол`,
        STR_STATUS_FALL: `Падение`,
        STR_STATUS_INCIDENT: `ДТП`,
        STR_STATUS_FINISHED: `Финишировал`,
        STR_STATUS_DNF: `Сход с дистанции`,
        STR_LIVE_MAP: `Интерактивная карта`,
    },
    }

    return data[conf.BotLang], nil
}
