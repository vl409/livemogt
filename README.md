**LiveMOGT:** Telegram bot for riders tracking
*Designed specially for the VCC MOGT series!*

How this works:

User: shares GEO location to telegram bot LiveMogt
User: visits https://example.com/livemogt to see all bot users on a map

server:
    1) livemogt (go daemon) - telegram bot accepting location and status
    2) webmap (go daemon) - HTTP server streaming users info to a webpage
    3) webpage (js app) - draws map with users, listens to events from webmap
    4) nginx - public endpoint, serves static files and proxies to webmap

         [in telegram]            [ on server]
         user1 ---> location ----> livemogt ---> {JSON updates} ---> webmap
         user2 ---> status   ---/                                   / /
         users ---> location --/                                   / /
                                                                  / /
         [ in browser ]                                          / /
         user1 ---> GET /livemogt ------------------------------/ /
              <------- map.js <--------- [ event source socket ]-/
              [ interactive map, receiving updates ]


**Configuration**
Edit files in conf, you have to provide:

 - Telegram bot Token
 - ID of telegram channel with your bot users
 - Public URL for the map
 - track.gpx file to be used

**Building**

    $ make

will run the make process inside docker. The result is a couple of fully static
go binaries with bot and web server and web assets ready to deploy into your
public website. Things can be also built locally without docker by invoking
corresponding makefiles in front/ and back/.




