
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

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/stat.h>
#include <signal.h>
#include <wait.h>
#include <time.h>

#define OK   0
#define ERR -1

#define TFMT "%Y/%m/%d %H:%M:%S"

#define mi_tmlog(fmt, args...)                                                \
    do {                                                                      \
        time_t t = time(NULL);                                                \
        struct tm *lt = localtime(&t);                                        \
        char buf[64];                                                         \
        if (strftime(buf, sizeof(buf), TFMT, lt) != 0) {                      \
            fprintf(stderr, "%s mi: ", buf);                                  \
        }                                                                     \
        fprintf(stderr, fmt, args);                                           \
    } while (0);

#define mi_log(fmt, args...) do { fprintf(stderr, fmt, args); } while (0)


typedef struct {
    const char   *exe;
    char        **argz;
    pid_t         pid;
} mi_child_t;


static mi_child_t    *procs;
static unsigned int   nprocs;


int
mi_init_proc(mi_child_t *proc, char *input)
{
    char          *p, *token;
    unsigned int   nargs, i;

    /* we start with non-empty input, i.e. program name at least */
    nargs = 1;

    for (p = input; *p; p++) {
        if (*p == ':') {
            nargs++;
        }
    }

    /* also allocate for terminating null-string */
    proc->argz = malloc(sizeof(char *) * (nargs + 1));
    if (proc->argz == NULL) {
        mi_tmlog("malloc failed: %s\n", strerror(errno));
        return ERR;
    }

    i = 0;
    p = input;

    while (1) {
        token = strsep(&p, ":");
        if (token == NULL) {
            break;
        }
        proc->argz[i++] = token;
    }

    proc->argz[i] = NULL;

    proc->exe = proc->argz[0];

    proc->pid = -1;

    return OK;
}


int
mi_notify_children(mi_child_t *procs, unsigned int nprocs, int signo)
{
    int           rc;
    unsigned int  i;

    rc = OK;

    for (i = 0; i < nprocs; i++) {
        if (procs[i].pid == -1) {
            continue;
        }

        if (kill(procs[i].pid, signo) == -1) {
            mi_tmlog("failed to send signal to pid %d\n", procs[i].pid);
            rc = ERR;
        }
    }

    return rc;
}


void
sigint_handler(int signo)
{
    (void) mi_notify_children(procs, nprocs, SIGTERM);
}


int
main(int argc, char *argv[], char *envp[])
{
    int           wstat, rv;
    pid_t         wpid;
    struct stat   st;
    unsigned int  i, nchild;

    rv = 0;

    if (argc < 2) {
        mi_log("Usage: %s <CMD> [CMD]...\n", argv[0]);
        mi_log("%s", "  CMD: /path/to/bin:arg1:arg2:...\n");
        mi_log("  Example: %s /bin/ls:-la /bin/ps:aux\n", argv[0]);
        mi_log("%s: no arguments provided, exiting\n", argv[0]);
        exit(EXIT_FAILURE);
    }

    nprocs = argc - 1;

    procs = malloc(nprocs * sizeof(mi_child_t));
    if (procs == NULL) {
        mi_tmlog("malloc() failed: %s\n", strerror(errno));
        exit(EXIT_FAILURE);
    }

    for (i = 0; i < nprocs; i++) {
        if (mi_init_proc(&procs[i], argv[i + 1]) != OK) {
            exit(EXIT_FAILURE);
        }

        if (stat(procs[i].exe, &st) != 0) {
            mi_tmlog("file does not exist: %s\n", procs[i].exe);
            exit(EXIT_FAILURE);
        }
    }

    nchild = 0;

    for (i = 0; i < nprocs; i++) {

        procs[i].pid = fork();

        if (procs[i].pid == -1) {
            mi_tmlog("fork failed(): %s\n", strerror(errno));

            /* send SIGTERM to all forked children */
            if (mi_notify_children(procs, nprocs, SIGTERM) != OK) {
                /*  we had problems with sending signals - just give up */
                mi_tmlog("%s", "hard fail occured, zombies expected\n");
                exit(EXIT_FAILURE);
            }

            rv = -1;

            goto wait_children;
        }

        if (procs[i].pid > 0) {
            /* parent */
            nchild++;

            mi_tmlog(">> child process '%s' started PID:%d\n",
                     procs[i].exe, procs[i].pid);

        } else {
            /* child */
            if (execve(procs[i].exe, procs[i].argz, envp) == -1) {
                mi_tmlog("execve failed: %s\n", strerror(errno));
                exit(EXIT_FAILURE);
            }

            /* does not return on success */
        }
    }

    if (signal(SIGINT, sigint_handler) == SIG_ERR) {
        mi_tmlog("failed to install SIGINT handler: %s\n", strerror(errno));
    }

wait_children:

    while (nchild) {

        wpid = wait(&wstat);

        if (wpid == -1) {

            if (errno == EINTR) {
                mi_tmlog("%s", ">> wait() interrupted by signal, ignored\n");
                continue;
            }

            mi_tmlog("wait() failed: %s, expect zombies", strerror(errno));
            exit(EXIT_FAILURE);
        }

        if (WIFEXITED(wstat) || WIFSIGNALED(wstat)) {

            mi_tmlog(">> child process pid:%d", wpid);

            if (WIFEXITED(wstat)) {
                mi_log(" exited, status=%d\n", WEXITSTATUS(wstat));

            } else {
                mi_log(" killed by signal %d\n", WTERMSIG(wstat));
            }

            nchild--;

            for (i = 0; i < nprocs; i++) {
                if (procs[i].pid == wpid) {
                    procs[i].pid = -1;
                    mi_tmlog(">> child process '%s' cleaned up\n", procs[i].exe);
                    break;
                }
            }
        }
    }

    mi_tmlog("%s", ">> all children complete, exiting\n");

    return rv;
}
