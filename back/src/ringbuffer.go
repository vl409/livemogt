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

type RingBuffer struct {
    buffer   []interface{}
    size     int
    index    int
}


func CreateRing(size int) *RingBuffer {
    return &RingBuffer{
        buffer: make([]interface{}, size),
        size: size,
        index: 0,
    }
}


func (rng *RingBuffer) push(data interface{}) {
    rng.buffer[rng.index] = data
    rng.index = (rng.index + 1) % rng.size
}


func (rng *RingBuffer) extract() []GeoPos {

    var values []GeoPos

    i := rng.index

    for _ = range rng.buffer {
        if rng.buffer[i] == nil {
            /* skip items not yet filled in */
            i = (i + 1) % rng.size
            continue
        }
        values = append(values, rng.buffer[i].(GeoPos))
        i = (i + 1) % rng.size
    }

    return values
}
