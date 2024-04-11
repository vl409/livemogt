
/*
 * Copyright (C) 2024 Vladimir Homutov
 */

/*
 * This file is part of microinit.
 *
 * microinit is free software: you can redistribute it and/or modify
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

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>
#include <time.h>


int main(int argc, char *argv[], char *envp[])
{
    char          **env;
    pid_t           pid;
    unsigned int    i;

    pid = getpid();

    printf("%d: argc: %d\n", pid, argc);

    for (i = 0; i < argc; i++) {
        printf("%d: argv[%d]='%s'\n", pid, i, argv[i]);
    }

    env = envp;
    i = 0;

    while (*env) {
        if (strcmp(*env, "MI_FOO=MI_BAR") == 0
            || strcmp(*env, "MI_BAR=MI_FOO") ==0)
        {
            printf("%d: env[%d]='%s'\n", pid, i, *env);
        }
        i++;
        env++;
    }

    for (i = 1; i < argc; i++) {

        if (strcmp(argv[i], "exit_1") == 0) {
            exit(1);
        }

        if (strcmp(argv[i], "exit_2") == 0) {
            exit(2);
        }

        if (strcmp(argv[i], "trap") == 0) {
            int z = 0;
            int d = 3 / z;
            return d;
        }

        if (strcmp(argv[i], "signal_int") == 0) {
            kill(pid, SIGINT);
            while (1);
        }

        if (strcmp(argv[i], "signal_kill") == 0) {
            kill(pid, SIGKILL);
            while (1);
        }
    }

    return 0;
}
