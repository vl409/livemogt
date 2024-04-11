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
    "time"
)

/* JSONs flowing between parties */

/* some data to send in keepalive message */
type KeepalivePing struct {
    Alive    bool
}

/* json position update */
type UserPosition struct {
    UserName string
    Lat      float64
    Lon      float64
    Last     time.Time
}

/* json status update */
type UserStatus struct {
    UserName     string
    Status       string
    MovingState  string
}
