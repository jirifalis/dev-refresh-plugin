# Chrome extension with websocket server for cmd line refresh of all localhost tabs

I was looking for a solution to update my browser from the command line, but all the solutions were outdated or didn't work in my specific environment. 

I use Fedora Linux with Hyprland tiling compositor. My usual workflow is to code on one workspace and have a web browser to test on the other workspace. So I want to always have an updated browser after switching no matter what project I'm working on.

So I ended with a simple browser extension and a simple websocket server. 

The browser extension automatically connects to the websocket server and waits for a refresh command. Which I can simply send from the command line while switching workspaces. It works great.

## Websocket server
It's written in Go and listens on a local IP. I did not use a websocket library with complex features, instead, I wrote my own with only what I needed.

## Browser extension
It's a chromium extension that connects to the server and waits. On failure or closed connection, it automatically reconnects. And there is an icon with a connection indicator, so I know if it's connected or not.

### Installation:
1. Open chromium
2. Go to `chrome://extensions/`
3. Enable "Developer mode"
4. Click "Load unpacked" and select a folder with the extension

## Build server:
``` 
go build dev-refresh-plugin-server.go 
```

## Usage:
The server is controlled via `dev-refresh-plugin-ctl` script.

```
dev-refresh-plugin-ctl start|stop|status|refresh
```

## Usage with Hyprland

In hyprland.conf

```
...
...

# start refresh server
exec-once=dev-refresh-plugin-ctl start

... 
...

# refresh browser
bind = $mainMod, 2, exec, dev-refresh-plugin-ctl refresh

...
...
```

